package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	"go.opentelemetry.io/otel"
)

// consumer listens for OrderCreatedEvents from RabbitMQ and initiates the payment process.
type consumer struct {
	service PaymentsService
}

func NewConsumer(service PaymentsService) *consumer {
	return &consumer{service}
}

// Listen connects to the OrderCreatedEvent queue and starts consuming messages continuously.
func (c *consumer) Listen(ch *amqp.Channel) {
	// Declare a durable queue specific to the payments service.
	q, err := ch.QueueDeclare("payments_"+broker.OrderCreatedEvent, true, false, false, false, amqp.Table{"x-dead-letter-exchange": "dlx_main"})
	if err != nil {
		log.Fatal(err)
	}

	// Bind the queue to the fanout exchange
	err = ch.QueueBind(q.Name, "", broker.OrderCreatedEvent, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	var forever chan struct{}

	// Process incoming messages in a background goroutine
	go func() {
		for d := range msgs {
			// Extract distributed tracing headers injected by the orders service
			ctx := broker.ExtractAMQPHeader(context.Background(), d.Headers)

			tr := otel.Tracer("amqp")
			_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - consume - %s", q.Name))

			o := &pb.Order{}
			if err := json.Unmarshal(d.Body, o); err != nil {
				// Reject immediately without requeueing for unparseable data (goes to DLQ)
				d.Nack(false, false)
				log.Printf("failed to unmarshal order: %v", err)
				continue
			}

			// Call the core business logic to create a payment link
			paymentLink, err := c.service.CreatePayment(ctx, o)
			if err != nil {
				log.Printf("failed to create payment: %v", err)

				// Push to a delay queue for exponential backoff retry.
				// This handles transient errors from the payment gateway.
				if err := broker.HandleRetry(ch, &d); err != nil {
					log.Printf("Error handling retry: %v", err)
				}

				d.Nack(false, false)

				continue
			}

			messageSpan.AddEvent(fmt.Sprintf("payment.created: %s", paymentLink))
			messageSpan.End()

			log.Printf("Payment link created %s", paymentLink)
			// Acknowledge the message so RabbitMQ removes it from the queue
			d.Ack(false)
		}
	}()

	<-forever
}
