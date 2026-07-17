package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
)

type PaymentsService interface {
	CreatePayment(context.Context, *pb.Order) (string, error)
}
