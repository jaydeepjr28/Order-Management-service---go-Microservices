package main

import (
	"context"
	"testing"

	"github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/omsv2-payments/gateway"
	"github.com/jaydeepjr28/omsv2-payments/processor/inmem"
)

type fakeOrdersGateway struct {
	orderID     string
	paymentLink string
}

func (g *fakeOrdersGateway) UpdateOrderAfterPaymentLink(_ context.Context, orderID, paymentLink string) error {
	g.orderID = orderID
	g.paymentLink = paymentLink
	return nil
}

var _ gateway.OrdersGateway = (*fakeOrdersGateway)(nil)

func TestService(t *testing.T) {
	processor := inmem.NewInmem()
	ordersGateway := &fakeOrdersGateway{}
	svc := NewService(processor, ordersGateway)

	t.Run("should create a payment link", func(t *testing.T) {
		order := &api.Order{ID: "order-1", CustomerID: "customer-1"}
		link, err := svc.CreatePayment(context.Background(), order)
		if err != nil {
			t.Errorf("CreatePayment() error = %v, want nil", err)
		}

		if link == "" {
			t.Error("CreatePayment() link is empty")
		}

		if ordersGateway.orderID != order.ID || ordersGateway.paymentLink != link {
			t.Error("CreatePayment() did not persist the generated payment link")
		}
	})
}
