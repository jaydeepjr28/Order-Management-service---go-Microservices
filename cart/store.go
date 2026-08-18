package main

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/redis/go-redis/v9"
)

type store struct {
	client *redis.Client
}

func NewRedisStore(addr string) *store {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &store{client: client}
}

func (s *store) AddItem(ctx context.Context, customerID string, item *pb.CartItem) error {
	cart, err := s.GetCart(ctx, customerID)

	if err != nil && err != redis.Nil {
		return err
	}

	if cart == nil {
		cart = &pb.Cart{
			CustomerID:	customerID,
			Items:		[]*pb.CartItem{},
		}
	}

	found := false
	for _, cartItem := range cart.Items {
		if cartItem.ProductID == item.ProductID {
			cartItem.Quantity += item.Quantity
			found = true
			break
		}
	}

	if !found {
		cart.Items = append(cart.Items, item)
	}

	return s.saveCart(ctx, cart)
}

func (s *store) RemoveItem(ctx context.Context, customerID, productID string) error {
	cart, err := s.GetCart(ctx, customerID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil
	}

	var newItems []*pb.CartItem
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}

	cart.Items = newItems
	return s.saveCart(ctx, cart)
}

func (s *store) GetCart(ctx context.Context, customerID string) (*pb.Cart, error) {
	val, err := s.client.Get(ctx, customerID).Result()
	if err == redis.Nil {
		return nil, nil
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

func (s *store) ClearCart(ctx context.Context, customerID string) error {
	return s.client.Del(ctx, customerID).Err()
}

func (s *store) saveCart(ctx context.Context, cart *pb.Cart) error {
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, cart.CustomerID, data, 24*time.Hour).Err()
}
