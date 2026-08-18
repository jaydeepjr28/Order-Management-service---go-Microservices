package inmem

import pb "github.com/jaydeepjr28/commons/api"

type Inmem struct{}

func NewInmem() *Inmem {
	return &Inmem{}
}

func (i *Inmem) CreatePaymentLink(order *pb.Order) (string, error) {
	return "/success.html?orderID=" + order.ID + "&customerID=" + order.CustomerID, nil
}
