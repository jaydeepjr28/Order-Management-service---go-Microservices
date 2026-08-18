package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jaydeepjr28/commons/broker"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn	*amqp.Connection
	ch	*amqp.Channel
}

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

	q, err := ch.QueueDeclare("notifications_queue", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(q.Name, "", broker.OrderCreatedEvent, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(q.Name, "", broker.UserRegisteredEvent, false, nil)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:	conn,
		ch:	ch,
	}, nil
}

func (c *Consumer) Listen() {
	msgs, err := c.ch.Consume("notifications_queue", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to register a consumer: %v", err)
	}

	for msg := range msgs {

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

func (c *Consumer) sendEmailStub(to, message string) {
	log.Printf("--> [STUB] Sending Email to %s: %s", to, message)
	time.Sleep(500 * time.Millisecond)
	log.Printf("--> [STUB] Email Sent successfully.")
}

func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
