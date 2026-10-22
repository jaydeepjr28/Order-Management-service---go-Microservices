package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
)

type StockService interface {
	CheckIfItemAreInStock(context.Context, []*pb.ItemsWithQuantity) (bool, []*pb.Item, error)
	GetItems(ctx context.Context, ids []string) ([]*pb.Item, error)
	DecrementItems(ctx context.Context, items []*pb.Item) error
	IncrementItems(ctx context.Context, items []*pb.Item) error
	CreateItem(ctx context.Context, item *pb.Item) (*pb.Item, error)
	UpdateItem(ctx context.Context, item *pb.Item) (*pb.Item, error)
}

type StockStore interface {
	GetItem(ctx context.Context, id string) (*pb.Item, error)
	GetItems(ctx context.Context, ids []string) ([]*pb.Item, error)
	DecrementItems(ctx context.Context, items []*pb.Item) error
	IncrementItems(ctx context.Context, items []*pb.Item) error
	CreateItem(ctx context.Context, item *pb.Item) (*pb.Item, error)
	UpdateItem(ctx context.Context, item *pb.Item) (*pb.Item, error)
}

