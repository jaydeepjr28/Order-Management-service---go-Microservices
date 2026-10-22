package main

import (
	"context"
	"encoding/json"
	"errors"

	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedOrderServiceServer

	service  OrdersService
	producer *broker.Producer
}

func NewGRPCHandler(grpcServer *grpc.Server, service OrdersService, producer *broker.Producer) {
	handler := &grpcHandler{
		service:  service,
		producer: producer,
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
	items, err := h.service.ValidateOrder(ctx, p)
	if err != nil {
		return nil, mapServiceError(err)
	}

	o, err := h.service.CreateOrder(ctx, p, items)
	if err != nil {
		return nil, mapServiceError(err)
	}

	marshalledOrder, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	if h.producer != nil {
		_ = h.producer.Publish(ctx, broker.OrderCreatedEvent, []byte(o.ID), marshalledOrder)
	}

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
