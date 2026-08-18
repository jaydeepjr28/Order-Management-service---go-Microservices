package common

import "errors"

// Common domain errors used across multiple microservices.
var (
	// ErrNoItems indicates that a cart or an order was submitted without any items.
	ErrNoItems = errors.New("items must have at least one item")
	
	// ErrNoStock indicates that an order could not be fulfilled because the requested quantity
	// of an item exceeds the available inventory in the stock service.
	ErrNoStock = errors.New("some item is not in stock")
)