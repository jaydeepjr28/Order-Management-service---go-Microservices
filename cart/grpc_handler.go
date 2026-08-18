package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedCartServiceServer
	store	*store
}

func NewGRPCHandler(grpcServer *grpc.Server, store *store) {
	handler := &grpcHandler{store: store}
	pb.RegisterCartServiceServer(grpcServer, handler)
}

func (h *grpcHandler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.Cart, error) {
	if err := h.store.AddItem(ctx, req.CustomerID, req.Item); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add item: %v", err)
	}

	return h.store.GetCart(ctx, req.CustomerID)
}

func (h *grpcHandler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.Cart, error) {
	if err := h.store.RemoveItem(ctx, req.CustomerID, req.ProductID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove item: %v", err)
	}

	return h.store.GetCart(ctx, req.CustomerID)
}

func (h *grpcHandler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.Cart, error) {
	cart, err := h.store.GetCart(ctx, req.CustomerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart: %v", err)
	}

	if cart == nil {
		return &pb.Cart{CustomerID: req.CustomerID, Items: []*pb.CartItem{}}, nil
	}

	return cart, nil
}

func (h *grpcHandler) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.Cart, error) {
	if err := h.store.ClearCart(ctx, req.CustomerID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to clear cart: %v", err)
	}

	return &pb.Cart{CustomerID: req.CustomerID, Items: []*pb.CartItem{}}, nil
}
