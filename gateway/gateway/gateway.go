package gateway

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
)

type OrdersGateway interface {
	CreateOrder(context.Context, *pb.CreateOrderRequest) (*pb.Order, error)
	GetOrder(ctx context.Context, orderID, customerID string) (*pb.Order, error)
	GetCustomerOrders(ctx context.Context, customerID string) (*pb.GetCustomerOrdersResponse, error)
	GetItems(ctx context.Context) ([]*pb.Item, error)
	CreateItem(ctx context.Context, item *pb.Item) (*pb.Item, error)
	UpdateItem(ctx context.Context, item *pb.Item) (*pb.Item, error)
}

type CartGateway interface {
	AddItem(context.Context, *pb.AddItemRequest) (*pb.Cart, error)
	RemoveItem(context.Context, *pb.RemoveItemRequest) (*pb.Cart, error)
	GetCart(context.Context, *pb.GetCartRequest) (*pb.Cart, error)
	ClearCart(context.Context, *pb.ClearCartRequest) (*pb.Cart, error)
}
