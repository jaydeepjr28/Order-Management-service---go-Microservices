package gateway

import (
	"context"
	"log"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/discovery"
)

type gateway struct {
	registry discovery.Registry
}

func NewGRPCGateway(registry discovery.Registry) *gateway {
	return &gateway{registry}
}

func (g *gateway) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {

	conn, err := discovery.ServiceConnection(context.Background(), "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewOrderServiceClient(conn)

	return c.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID:	p.CustomerID,
		Items:		p.Items,
	})
}

func (g *gateway) GetOrder(ctx context.Context, orderID, customerID string) (*pb.Order, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewOrderServiceClient(conn)

	return c.GetOrder(ctx, &pb.GetOrderRequest{
		OrderID:	orderID,
		CustomerID:	customerID,
	})
}

func (g *gateway) GetCustomerOrders(ctx context.Context, customerID string) (*pb.GetCustomerOrdersResponse, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewOrderServiceClient(conn)

	return c.GetCustomerOrders(ctx, &pb.GetCustomerOrdersRequest{
		CustomerID: customerID,
	})
}

func (g *gateway) GetItems(ctx context.Context) ([]*pb.Item, error) {

	conn, err := discovery.ServiceConnection(context.Background(), "stock", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewStockServiceClient(conn)

	res, err := c.GetItems(ctx, &pb.GetItemsRequest{})
	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

func (g *gateway) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.AddItem(ctx, req)
}

func (g *gateway) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.RemoveItem(ctx, req)
}

func (g *gateway) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.GetCart(ctx, req)
}

func (g *gateway) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.ClearCart(ctx, req)
}
