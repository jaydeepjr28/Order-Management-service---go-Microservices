package main

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/jaydeepjr28/commons/api"
)

type Store struct {
	mu	sync.Mutex
	stock	map[string]*pb.Item
}

func NewStore() *Store {
	return &Store{
		stock: map[string]*pb.Item{
			"2": {
				ID:		"2",
				Name:		"Potato Chips",
				PriceID:	"price_1PBaC2RoHcnwnsLnYLWpx1CK",
				Quantity:	10,
			},
			"1": {
				ID:		"1",
				Name:		"Cheese Burger",
				PriceID:	"price_1PA7xYRoHcnwnsLn4CrfWAvG",
				Quantity:	20,
			},
		},
	}
}

func (s *Store) GetItem(ctx context.Context, id string) (*pb.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.stock {
		if item.ID == id {
			return item, nil
		}
	}

	return nil, fmt.Errorf("item not found")
}

func (s *Store) GetItems(ctx context.Context, ids []string) ([]*pb.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res []*pb.Item
	if len(ids) == 0 {

		for _, i := range s.stock {
			res = append(res, i)
		}
		return res, nil
	}

	for _, id := range ids {
		if i, ok := s.stock[id]; ok {
			res = append(res, i)
		}
	}

	return res, nil
}

func (s *Store) DecrementItems(ctx context.Context, items []*pb.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		stockItem, ok := s.stock[item.ID]
		if !ok {
			return fmt.Errorf("item not found: %s", item.ID)
		}
		if stockItem.Quantity < item.Quantity {
			return fmt.Errorf("insufficient stock for item %s", item.ID)
		}
	}

	for _, item := range items {
		stockItem := s.stock[item.ID]
		stockItem.Quantity -= item.Quantity
	}

	return nil
}
