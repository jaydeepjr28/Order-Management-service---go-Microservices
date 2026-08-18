package gateway

import (
	"context"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/discovery"
)

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry}
}

func (g *Gateway) GetOrder(ctx context.Context, orderID, customerID string) (*pb.Order, error) {
	conn, err := discovery.ServiceConnection(ctx, "orders", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)
	return client.GetOrder(ctx, &pb.GetOrderRequest{
		OrderID:    orderID,
		CustomerID: customerID,
	})
}
