package main

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/jaydeepjr28/commons/api"
)

// Store is an in-memory implementation of the stock database.
// In a real application, this would be a persistent database like PostgreSQL or Redis.
type Store struct {
	mu    sync.Mutex          // Protects concurrent map access
	stock map[string]*pb.Item // Map of itemID to Item details
}

// NewStore initializes the in-memory catalog with some default products.
// These are the products that will be displayed on the frontend.
func NewStore() *Store {
	return &Store{
		stock: map[string]*pb.Item{
			"2": {
				ID:       "2",
				Name:     "Potato Chips",
				PriceID:  "price_1PBaC2RoHcnwnsLnYLWpx1CK",
				Quantity: 10,
			},
			"1": {
				ID:       "1",
				Name:     "Cheese Burger",
				PriceID:  "price_1PA7xYRoHcnwnsLn4CrfWAvG",
				Quantity: 20,
			},
		},
	}
}

// GetItem fetches a single item by ID from the in-memory map.
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

// GetItems fetches multiple items by their IDs.
// If the requested IDs slice is empty, it returns the entire catalog.
func (s *Store) GetItems(ctx context.Context, ids []string) ([]*pb.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res []*pb.Item
	if len(ids) == 0 {
		// Return all items in the store
		for _, i := range s.stock {
			res = append(res, i)
		}
		return res, nil
	}
	// Return only the requested items
	for _, id := range ids {
		if i, ok := s.stock[id]; ok {
			res = append(res, i)
		}
	}

	return res, nil
}

// DecrementItems reduces the available stock quantity for a list of purchased items.
// This is typically called after a successful payment.
func (s *Store) DecrementItems(ctx context.Context, items []*pb.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Verify we have enough stock for all items before making any deductions
	for _, item := range items {
		stockItem, ok := s.stock[item.ID]
		if !ok {
			return fmt.Errorf("item not found: %s", item.ID)
		}
		if stockItem.Quantity < item.Quantity {
			return fmt.Errorf("insufficient stock for item %s", item.ID)
		}
	}

	// 2. Perform the actual deductions
	for _, item := range items {
		stockItem := s.stock[item.ID]
		stockItem.Quantity -= item.Quantity
	}

	return nil
}
