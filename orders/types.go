package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
)

type OrdersService interface {
	CreateOrder(context.Context, *pb.CreateOrderRequest, []*pb.Item) (*pb.Order, error)
	ValidateOrder(context.Context, *pb.CreateOrderRequest) ([]*pb.Item, error)
	GetOrder(context.Context, *pb.GetOrderRequest) (*pb.Order, error)
	GetCustomerOrders(context.Context, *pb.GetCustomerOrdersRequest) (*pb.GetCustomerOrdersResponse, error)
	UpdateOrder(context.Context, *pb.Order) (*pb.Order, error)
}

type OrdersStore interface {
	Create(context.Context, Order) (string, error)
	Get(ctx context.Context, id, customerID string) (*Order, error)
	GetByCustomer(ctx context.Context, customerID string) ([]*Order, error)
	GetAll(ctx context.Context) ([]*Order, error)
	Update(ctx context.Context, id string, o *pb.Order) error
}

type Order struct {
	ID		string		`json:"id"`
	CustomerID	string		`json:"customerID"`
	Status		string		`json:"status"`
	PaymentLink	string		`json:"paymentLink"`
	Items		[]*pb.Item	`json:"items"`
}

func (o *Order) ToProto() *pb.Order {
	return &pb.Order{
		ID:		o.ID,
		CustomerID:	o.CustomerID,
		Status:		o.Status,
		PaymentLink:	o.PaymentLink,
		Items:		o.Items,
	}
}
