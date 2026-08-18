package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/omsv2-payments/gateway"
	"github.com/jaydeepjr28/omsv2-payments/processor"
)

// service implements the core PaymentsService logic.
// It orchestrates interactions between the payment processor (e.g. Razorpay) and the downstream Orders gateway.
type service struct {
	processor processor.PaymentProcessor
	gateway   gateway.OrdersGateway
}

func NewService(processor processor.PaymentProcessor, gateway gateway.OrdersGateway) *service {
	return &service{processor, gateway}
}

// CreatePayment generates a payment link for a pending order and updates the order with that link.
func (s *service) CreatePayment(ctx context.Context, o *pb.Order) (string, error) {
	// Call the configured processor (Razorpay or inmem mock) to generate a hosted checkout URL
	link, err := s.processor.CreatePaymentLink(o)
	if err != nil {
		return "", err
	}

	// Persist the generated link in the 'orders' microservice database via gRPC
	err = s.gateway.UpdateOrderAfterPaymentLink(ctx, o.ID, link)
	if err != nil {
		return "", err
	}

	return link, nil
}
