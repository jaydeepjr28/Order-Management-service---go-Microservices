package main

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/redis/go-redis/v9"
)

// store encapsulates the Redis client used for the Cart service.
// Carts are ephemeral, so Redis is an ideal in-memory key-value store for this data.
type store struct {
	client *redis.Client
}

func NewRedisStore(addr string) *store {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &store{client: client}
}

// AddItem retrieves the current cart, modifies it by adding/incrementing the requested item,
// and saves the updated cart back to Redis.
func (s *store) AddItem(ctx context.Context, customerID string, item *pb.CartItem) error {
	cart, err := s.GetCart(ctx, customerID)
	// redis.Nil means the key doesn't exist, which is fine for a new cart.
	if err != nil && err != redis.Nil {
		return err
	}

	if cart == nil {
		cart = &pb.Cart{
			CustomerID: customerID,
			Items:      []*pb.CartItem{},
		}
	}

	// Check if the item is already in the cart to increment its quantity
	found := false
	for _, cartItem := range cart.Items {
		if cartItem.ProductID == item.ProductID {
			cartItem.Quantity += item.Quantity
			found = true
			break
		}
	}

	// If not found, add it as a new line item
	if !found {
		cart.Items = append(cart.Items, item)
	}

	return s.saveCart(ctx, cart)
}

// RemoveItem completely removes a line item from the cart, regardless of its quantity.
func (s *store) RemoveItem(ctx context.Context, customerID, productID string) error {
	cart, err := s.GetCart(ctx, customerID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil // Nothing to remove if cart doesn't exist
	}

	// Filter out the item to be removed
	var newItems []*pb.CartItem
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}

	cart.Items = newItems
	return s.saveCart(ctx, cart)
}

// GetCart retrieves the Cart protobuf object from Redis.
func (s *store) GetCart(ctx context.Context, customerID string) (*pb.Cart, error) {
	val, err := s.client.Get(ctx, customerID).Result()
	if err == redis.Nil {
		return nil, nil // Key doesn't exist
	}
	if err != nil {
		return nil, err
	}

	var cart pb.Cart
	if err := json.Unmarshal([]byte(val), &cart); err != nil {
		return nil, err
	}

	return &cart, nil
}

// ClearCart deletes the cart key from Redis.
func (s *store) ClearCart(ctx context.Context, customerID string) error {
	return s.client.Del(ctx, customerID).Err()
}

// saveCart serializes the Cart protobuf to JSON and stores it in Redis with a 24-hour TTL.
func (s *store) saveCart(ctx context.Context, cart *pb.Cart) error {
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	// Expire carts after 24 hours of inactivity
	return s.client.Set(ctx, cart.CustomerID, data, 24*time.Hour).Err()
}
