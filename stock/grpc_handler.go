package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StockGrpcHandler struct {
	pb.UnimplementedStockServiceServer

	service StockService
}

func NewGRPCHandler(
	server *grpc.Server,
	stockService StockService,
) {
	handler := &StockGrpcHandler{
		service: stockService,
	}

	pb.RegisterStockServiceServer(server, handler)
}

func (s *StockGrpcHandler) CheckIfItemIsInStock(ctx context.Context, p *pb.CheckIfItemIsInStockRequest) (*pb.CheckIfItemIsInStockResponse, error) {
	inStock, items, err := s.service.CheckIfItemAreInStock(ctx, p.Items)
	if err != nil {
		return nil, err
	}

	return &pb.CheckIfItemIsInStockResponse{
		InStock:	inStock,
		Items:		items,
	}, nil
}

func (s *StockGrpcHandler) GetItems(ctx context.Context, payload *pb.GetItemsRequest) (*pb.GetItemsResponse, error) {
	items, err := s.service.GetItems(ctx, payload.ItemIDs)
	if err != nil {
		return nil, err
	}

	return &pb.GetItemsResponse{
		Items: items,
	}, nil
}

func (s *StockGrpcHandler) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.Item, error) {
	if req == nil || req.Item == nil {
		return nil, status.Errorf(codes.InvalidArgument, "item is required")
	}
	return s.service.CreateItem(ctx, req.Item)
}

func (s *StockGrpcHandler) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.Item, error) {
	if req == nil || req.Item == nil {
		return nil, status.Errorf(codes.InvalidArgument, "item is required")
	}
	return s.service.UpdateItem(ctx, req.Item)
}
