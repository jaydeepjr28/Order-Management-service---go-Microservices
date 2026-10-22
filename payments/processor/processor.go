package processor

import pb "github.com/jaydeepjr28/commons/api"

type PaymentProcessor interface {
	CreatePaymentLink(*pb.Order) (string, error)
	RefundPayment(*pb.Order) error
}

