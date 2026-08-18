package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
)

// Service implements the core business logic for the Stock domain.
// It interacts with the local datastore to check and decrement product inventory.
type Service struct {
	store StockStore
}

func NewService(store StockStore) *Service {
	return &Service{store}
}

// CheckIfItemAreInStock verifies if the requested quantities for a list of items are available.
// If all items are in stock, it returns true along with the detailed Item structs (including prices).
// If any item is out of stock, it returns false.
func (s *Service) CheckIfItemAreInStock(ctx context.Context, p []*pb.ItemsWithQuantity) (bool, []*pb.Item, error) {
	itemIDs := make([]string, 0)
	for _, item := range p {
		itemIDs = append(itemIDs, item.ID)
	}

	// Fetch current stock levels for the requested items
	itemsInStock, err := s.store.GetItems(ctx, itemIDs)
	if err != nil {
		return false, nil, err
	}

	// Validate that we have enough stock for each requested item
	for _, stockItem := range itemsInStock {
		for _, reqItem := range p {
			if stockItem.ID == reqItem.ID && stockItem.Quantity < reqItem.Quantity {
				return false, itemsInStock, nil
			}
		}
	}

	// Build the response with the exact requested quantities and current prices
	items := make([]*pb.Item, 0)
	for _, stockItem := range itemsInStock {
		for _, reqItem := range p {
			if stockItem.ID == reqItem.ID {
				items = append(items, &pb.Item{
					ID:       stockItem.ID,
					Name:     stockItem.Name,
					PriceID:  stockItem.PriceID,
					Quantity: reqItem.Quantity,
				})
			}
		}
	}

	return true, items, nil
}

// GetItems retrieves the full details for a list of item IDs.
// Passing an empty slice retrieves all items (used for the frontend catalog).
func (s *Service) GetItems(ctx context.Context, ids []string) ([]*pb.Item, error) {
	return s.store.GetItems(ctx, ids)
}

// DecrementItems deducts the purchased quantities from the available inventory.
func (s *Service) DecrementItems(ctx context.Context, items []*pb.Item) error {
	return s.store.DecrementItems(ctx, items)
}
