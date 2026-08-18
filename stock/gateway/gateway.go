package gateway

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
)

type OrdersGateway interface {
	GetOrder(ctx context.Context, orderID, customerID string) (*pb.Order, error)
}
