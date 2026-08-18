package broker

// Event routing keys used for publishing and consuming messages in RabbitMQ.
const (
	// OrderCreatedEvent is published by the orders service when a new order is successfully saved.
	// Services like payments listen to this to initiate the payment process.
	OrderCreatedEvent = "order.created"
	
	// OrderPaidEvent is published by the payments service when a payment is verified as successful.
	// Services like kitchen and notification listen to this to begin preparation and notify the user.
	OrderPaidEvent    = "order.paid"

	// UserRegisteredEvent is published by the gateway service when a new user signs up.
	// Services like notification listen to this to send a welcome email.
	UserRegisteredEvent = "user.registered"
)
