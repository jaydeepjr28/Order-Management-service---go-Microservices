package gateway

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
)

type KitchenGateway interface {
	UpdateOrder(context.Context, *pb.Order) error
}
