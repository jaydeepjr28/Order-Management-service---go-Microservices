package main

import (
	"context"

	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/omsv2-orders/gateway"
)

type service struct {
	store	OrdersStore
	gateway	gateway.StockGateway
}

func NewService(store OrdersStore, gateway gateway.StockGateway) *service {
	return &service{store, gateway}
}

func (s *service) GetOrder(ctx context.Context, p *pb.GetOrderRequest) (*pb.Order, error) {
	o, err := s.store.Get(ctx, p.OrderID, p.CustomerID)
	if err != nil {
		return nil, err
	}

	return o.ToProto(), nil
}

func (s *service) GetCustomerOrders(ctx context.Context, p *pb.GetCustomerOrdersRequest) (*pb.GetCustomerOrdersResponse, error) {
	orders, err := s.store.GetByCustomer(ctx, p.CustomerID)
	if err != nil {
		return nil, err
	}

	protoOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		protoOrders = append(protoOrders, o.ToProto())
	}

	return &pb.GetCustomerOrdersResponse{Orders: protoOrders}, nil
}

func (s *service) UpdateOrder(ctx context.Context, o *pb.Order) (*pb.Order, error) {
	err := s.store.Update(ctx, o.ID, o)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (s *service) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest, items []*pb.Item) (*pb.Order, error) {
	id, err := s.store.Create(ctx, Order{
		CustomerID:	p.CustomerID,
		Status:		"pending",
		Items:		items,
		PaymentLink:	"",
	})
	if err != nil {
		return nil, err
	}

	o := &pb.Order{
		ID:		id,
		CustomerID:	p.CustomerID,
		Status:		"pending",
		Items:		items,
	}

	return o, nil
}

func (s *service) ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) ([]*pb.Item, error) {
	if len(p.Items) == 0 {
		return nil, common.ErrNoItems
	}

	mergedItems := mergeItemsQuantities(p.Items)

	inStock, items, err := s.gateway.CheckIfItemIsInStock(ctx, p.CustomerID, mergedItems)
	if err != nil {
		return nil, err
	}
	if !inStock {
		return items, common.ErrNoStock
	}

	return items, nil
}

func mergeItemsQuantities(items []*pb.ItemsWithQuantity) []*pb.ItemsWithQuantity {
	merged := make([]*pb.ItemsWithQuantity, 0)

	for _, item := range items {
		found := false
		for _, finalItem := range merged {
			if finalItem.ID == item.ID {
				finalItem.Quantity += item.Quantity
				found = true
				break
			}
		}

		if !found {
			merged = append(merged, item)
		}
	}

	return merged
}
