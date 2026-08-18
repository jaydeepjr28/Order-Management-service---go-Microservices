package main

import (
	"context"

	pb "github.com/jaydeepjr28/commons/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcHandler implements the generated pb.CartServiceServer interface.
type grpcHandler struct {
	pb.UnimplementedCartServiceServer
	store *store
}

func NewGRPCHandler(grpcServer *grpc.Server, store *store) {
	handler := &grpcHandler{store: store}
	pb.RegisterCartServiceServer(grpcServer, handler)
}

// AddItem adds a product to the user's cart (or increments its quantity if it already exists).
// It returns the updated Cart.
func (h *grpcHandler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.Cart, error) {
	if err := h.store.AddItem(ctx, req.CustomerID, req.Item); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add item: %v", err)
	}

	return h.store.GetCart(ctx, req.CustomerID)
}

// RemoveItem removes a product entirely from the user's cart, regardless of quantity.
func (h *grpcHandler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.Cart, error) {
	if err := h.store.RemoveItem(ctx, req.CustomerID, req.ProductID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove item: %v", err)
	}

	return h.store.GetCart(ctx, req.CustomerID)
}

// GetCart fetches the user's current cart.
// If the cart doesn't exist, it returns an empty cart instead of an error.
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

// ClearCart deletes the user's cart (e.g. after a successful checkout).
func (h *grpcHandler) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.Cart, error) {
	if err := h.store.ClearCart(ctx, req.CustomerID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to clear cart: %v", err)
	}

	return &pb.Cart{CustomerID: req.CustomerID, Items: []*pb.CartItem{}}, nil
}
