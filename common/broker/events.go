package broker

import "time"

const (
	OrderCreatedEvent     = "order.created"
	OrderPaidEvent        = "order.paid"
	UserRegisteredEvent   = "user.registered"
	UserOTPRequestedEvent = "user.otp_requested"

	// Saga Pattern Events
	OrderFailedEvent      = "order.failed"
	OrderCancelledEvent   = "order.cancelled"
	OrderRefundedEvent    = "order.refunded"
	OrderStockFailedEvent = "order.stock_failed"
)

type UserOTPRequestedPayload struct {
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	OTP              string    `json:"otp"`
	ExpiresInSeconds int       `json:"expiresInSeconds"`
	RequestedAt      time.Time `json:"requestedAt"`
}

type UserRegisteredPayload struct {
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	CustomerID   string    `json:"customerID"`
	Role         string    `json:"role"`
	Provider     string    `json:"provider,omitempty"`
	RegisteredAt time.Time `json:"registeredAt"`
}


