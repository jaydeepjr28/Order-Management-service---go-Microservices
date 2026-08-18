package main

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/jaydeepjr28/commons/broker"
)

// Consumer manages the RabbitMQ connection and channel for the Notification service.
type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewConsumer connects to RabbitMQ, sets up the topic exchange, and creates a durable queue
// bound to specific routing keys (e.g. order creation, user registration).
func NewConsumer(addr string) (*Consumer, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s/", addr))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare a durable queue for notifications
	q, err := ch.QueueDeclare("notifications_queue", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	// Bind the queue to listen for order creation events (fanout)
	err = ch.QueueBind(q.Name, "", broker.OrderCreatedEvent, false, nil)
	if err != nil {
		return nil, err
	}

	// Bind the queue to listen for user registration events (fanout)
	err = ch.QueueBind(q.Name, "", broker.UserRegisteredEvent, false, nil)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn: conn,
		ch:   ch,
	}, nil
}

// Listen starts consuming messages from the notifications queue.
// Since auto-ack is set to true, messages are acknowledged immediately upon receipt.
func (c *Consumer) Listen() {
	msgs, err := c.ch.Consume("notifications_queue", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to register a consumer: %v", err)
	}

	// Process messages continuously
	for msg := range msgs {
		// Because we bind the queue to multiple fanout exchanges, we can use msg.Exchange to identify the event
		switch msg.Exchange {
		case broker.OrderCreatedEvent:
			log.Printf("[Notification Service] Sending Email/SMS for Order Created! Data: %s", string(msg.Body))
			c.sendEmailStub("customer@example.com", "Your order has been created!")
		case broker.UserRegisteredEvent:
			log.Printf("[Notification Service] Sending Welcome Email! Data: %s", string(msg.Body))
			c.sendEmailStub("newuser@example.com", "Welcome to the OMS platform!")
		default:
			log.Printf("[Notification Service] Unknown event received from exchange: %s", msg.Exchange)
		}
	}
}

// sendEmailStub simulates sending an email/SMS using a third-party API (e.g. Twilio, SendGrid)
func (c *Consumer) sendEmailStub(to, message string) {
	log.Printf("--> [STUB] Sending Email to %s: %s", to, message)
	time.Sleep(500 * time.Millisecond) // Simulate network latency
	log.Printf("--> [STUB] Email Sent successfully.")
}

// Close cleans up the AMQP channel and connection on shutdown.
func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
