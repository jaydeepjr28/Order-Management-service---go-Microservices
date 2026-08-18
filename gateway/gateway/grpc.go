package gateway

import (
	"context"
	"log"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/discovery"
)

// gateway implements the OrdersGateway and CartGateway interfaces.
// It acts as a gRPC client wrapper to communicate with downstream microservices.
type gateway struct {
	registry discovery.Registry
}

// NewGRPCGateway initializes a new gRPC gateway client using the provided service registry.
func NewGRPCGateway(registry discovery.Registry) *gateway {
	return &gateway{registry}
}

// CreateOrder forwards a checkout request to the 'orders' microservice.
func (g *gateway) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {
	// Look up the IP address of an active 'orders' service from Consul
	conn, err := discovery.ServiceConnection(context.Background(), "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	// Initialize the generated gRPC client for the OrderService
	c := pb.NewOrderServiceClient(conn)

	// Make the RPC call
	return c.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: p.CustomerID,
		Items:      p.Items,
	})
}

// GetOrder forwards a request to fetch an order's details to the 'orders' microservice.
func (g *gateway) GetOrder(ctx context.Context, orderID, customerID string) (*pb.Order, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewOrderServiceClient(conn)

	return c.GetOrder(ctx, &pb.GetOrderRequest{
		OrderID:    orderID,
		CustomerID: customerID,
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

// GetItems forwards a request to fetch all available products to the 'stock' microservice.
func (g *gateway) GetItems(ctx context.Context) ([]*pb.Item, error) {
	// Discover and connect to the 'stock' service
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


// AddItem forwards a request to add an item to the shopping cart to the 'cart' microservice.
func (g *gateway) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.AddItem(ctx, req)
}

// RemoveItem forwards a request to remove an item from the shopping cart to the 'cart' microservice.
func (g *gateway) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.RemoveItem(ctx, req)
}

// GetCart forwards a request to retrieve the current shopping cart state to the 'cart' microservice.
func (g *gateway) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.GetCart(ctx, req)
}

// ClearCart forwards a request to empty the shopping cart to the 'cart' microservice.
func (g *gateway) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.Cart, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "cart", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewCartServiceClient(conn)
	return c.ClearCart(ctx, req)
}
