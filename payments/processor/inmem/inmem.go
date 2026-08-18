package inmem

import pb "github.com/jaydeepjr28/commons/api"

// Inmem is a mock payment processor used for local development and testing
// when a valid Razorpay API key is not available.
type Inmem struct {}

func NewInmem() *Inmem {
	return &Inmem{}
}

// CreatePaymentLink generates a fake, immediate success URL that routes the user
// directly back to the frontend's success page without actually charging them.
func (i *Inmem) CreatePaymentLink(order *pb.Order) (string, error) {
	return "/success.html?orderID=" + order.ID + "&customerID=" + order.CustomerID, nil
}