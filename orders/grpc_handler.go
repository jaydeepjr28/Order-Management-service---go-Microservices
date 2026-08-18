package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	common "github.com/jaydeepjr28/commons"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcHandler implements the generated pb.OrderServiceServer interface.
// It acts as the transport layer, translating gRPC requests into calls to the core OrdersService.
type grpcHandler struct {
	pb.UnimplementedOrderServiceServer

	service OrdersService
	channel *amqp.Channel // RabbitMQ channel used to publish events
}

func NewGRPCHandler(grpcServer *grpc.Server, service OrdersService, channel *amqp.Channel) {
	handler := &grpcHandler{
		service: service,
		channel: channel,
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

// CreateOrder is the entrypoint for submitting a new order.
// It coordinates with the business layer to validate stock, saves the order in "pending" status,
// and then asynchronously publishes an OrderCreatedEvent to RabbitMQ so the payment service can process it.
func (h *grpcHandler) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {
	// Initialize tracing for the message publication phase
	tr := otel.Tracer("amqp")
	amqpContext, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - publish - %s", broker.OrderCreatedEvent))
	defer messageSpan.End()

	// 1. Validate: Contact the stock service to ensure we have enough inventory
	items, err := h.service.ValidateOrder(amqpContext, p)
	if err != nil {
		return nil, mapServiceError(err)
	}

	// 2. Create: Persist the new order in MongoDB as "pending"
	o, err := h.service.CreateOrder(amqpContext, p, items)
	if err != nil {
		return nil, mapServiceError(err)
	}

	marshalledOrder, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	// Inject the current trace headers into the AMQP message so the trace continues in the consumer service
	headers := broker.InjectAMQPHeaders(amqpContext)

	// 3. Publish: Send the OrderCreatedEvent to RabbitMQ's fanout exchange
	h.channel.PublishWithContext(amqpContext, broker.OrderCreatedEvent, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         marshalledOrder,
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
	})

	return o, nil
}

// mapServiceError translates core domain errors (like Out of Stock) into standard gRPC status codes.
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
