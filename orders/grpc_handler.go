package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedOrderServiceServer

	service	OrdersService
	channel	*amqp.Channel
}

func NewGRPCHandler(grpcServer *grpc.Server, service OrdersService, channel *amqp.Channel) {
	handler := &grpcHandler{
		service:	service,
		channel:	channel,
	}
	pb.RegisterOrderServiceServer(grpcServer, handler)
}

func (h *grpcHandler) UpdateOrder(ctx context.Context, p *pb.Order) (*pb.Order, error) {
	return h.service.UpdateOrder(ctx, p)
}

func (h *grpcHandler) GetOrder(ctx context.Context, p *pb.GetOrderRequest) (*pb.Order, error) {
	return h.service.GetOrder(ctx, p)
}

func (h *grpcHandler) GetCustomerOrders(ctx context.Context, p *pb.GetCustomerOrdersRequest) (*pb.GetCustomerOrdersResponse, error) {
	return h.service.GetCustomerOrders(ctx, p)
}

func (h *grpcHandler) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {

	tr := otel.Tracer("amqp")
	amqpContext, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - publish - %s", broker.OrderCreatedEvent))
	defer messageSpan.End()

	items, err := h.service.ValidateOrder(amqpContext, p)
	if err != nil {
		return nil, mapServiceError(err)
	}

	o, err := h.service.CreateOrder(amqpContext, p, items)
	if err != nil {
		return nil, mapServiceError(err)
	}

	marshalledOrder, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	headers := broker.InjectAMQPHeaders(amqpContext)

	h.channel.PublishWithContext(amqpContext, broker.OrderCreatedEvent, "", false, false, amqp.Publishing{
		ContentType:	"application/json",
		Body:		marshalledOrder,
		DeliveryMode:	amqp.Persistent,
		Headers:	headers,
	})

	return o, nil
}

func mapServiceError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, common.ErrNoItems), errors.Is(err, common.ErrNoStock):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return err
	}
}
