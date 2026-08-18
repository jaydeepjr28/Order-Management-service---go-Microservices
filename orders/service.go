package main

import (
	"context"

	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/omsv2-orders/gateway"
)

// service contains the core business logic for the Orders domain.
// It orchestrates interactions between the database (store) and external dependencies (stock gateway).
type service struct {
	store   OrdersStore
	gateway gateway.StockGateway
}

func NewService(store OrdersStore, gateway gateway.StockGateway) *service {
	return &service{store, gateway}
}

// GetOrder retrieves a specific order from the database and converts it to a gRPC protobuf struct.
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

// UpdateOrder modifies an existing order in the database.
// This is typically called by background workers (e.g. after receiving a payment success event).
func (s *service) UpdateOrder(ctx context.Context, o *pb.Order) (*pb.Order, error) {
	err := s.store.Update(ctx, o.ID, o)
	if err != nil {
		return nil, err
	}

	return o, nil
}

// CreateOrder initializes a new order document in MongoDB with a 'pending' status.
// The payment link is left empty until the downstream payment service generates it.
func (s *service) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest, items []*pb.Item) (*pb.Order, error) {
	id, err := s.store.Create(ctx, Order{
		CustomerID:  p.CustomerID,
		Status:      "pending",
		Items:       items,
		PaymentLink: "",
	})
	if err != nil {
		return nil, err
	}

	o := &pb.Order{
		ID:         id,
		CustomerID: p.CustomerID,
		Status:     "pending",
		Items:      items,
	}

	return o, nil
}

// ValidateOrder ensures the requested items are actually in stock.
// It merges duplicate items in the cart to calculate total quantity needed per item,
// then calls the downstream Stock service to verify availability.
func (s *service) ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) ([]*pb.Item, error) {
	if len(p.Items) == 0 {
		return nil, common.ErrNoItems
	}

	// Consolidate duplicates so we don't ask the stock service for the same item twice
	mergedItems := mergeItemsQuantities(p.Items)

	// Call the stock microservice to verify inventory
	inStock, items, err := s.gateway.CheckIfItemIsInStock(ctx, p.CustomerID, mergedItems)
	if err != nil {
		return nil, err
	}
	if !inStock {
		return items, common.ErrNoStock
	}

	return items, nil
}

// mergeItemsQuantities combines duplicate line items into a single line item with a summed quantity.
// e.g. [{id:1, qty:1}, {id:1, qty:2}] -> [{id:1, qty:3}]
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
