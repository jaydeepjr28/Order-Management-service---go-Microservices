package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// consumer listens for asynchronous events from RabbitMQ and updates the local order state.
type consumer struct {
	service OrdersService
}

func NewConsumer(service OrdersService) *consumer {
	return &consumer{service}
}

// Listen sets up the RabbitMQ queue, binds it to the OrderPaidEvent exchange, and starts consuming messages.
func (c *consumer) Listen(ch *amqp.Channel) {
	// Declare an anonymous queue that will automatically delete when the connection closes.
	// We configure a Dead Letter Exchange (dlx_main) so failed messages aren't lost.
	q, err := ch.QueueDeclare("", true, false, true, false, amqp.Table{"x-dead-letter-exchange": "dlx_main"})
	if err != nil {
		log.Fatal(err)
	}

	// Bind the queue to the OrderPaidEvent exchange. This ensures this service receives payment success events.
	err = ch.QueueBind(q.Name, "", broker.OrderPaidEvent, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	var forever chan struct{}

	// Start consuming messages in a background goroutine
	go func() {
		for d := range msgs {
			log.Printf("Received message: %s", d.Body)

			// Extract the OpenTelemetry trace headers injected by the publisher (payments service)
			ctx := broker.ExtractAMQPHeader(context.Background(), d.Headers)

			tr := otel.Tracer("amqp")
			_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - consume - %s", q.Name))

			o := &pb.Order{}
			if err := json.Unmarshal(d.Body, o); err != nil {
				// Reject unprocessable messages without requeueing, they will go to the DLQ
				d.Nack(false, false)
				log.Printf("failed to unmarshal order: %v", err)
				continue
			}

			// Update the local database with the new order state (e.g., status="paid")
			_, err := c.service.UpdateOrder(ctx, o)
			if err != nil {
				log.Printf("failed to update order: %v", err)

				// If database update fails, delegate to the exponential backoff retry handler
				if err := broker.HandleRetry(ch, &d); err != nil {
					log.Printf("Error handling retry: %v", err)
				}

				continue
			}

			messageSpan.AddEvent("order.updated")
			messageSpan.End()

			log.Println("Order has been updated from AMQP")
			// Acknowledge the message so RabbitMQ removes it from the queue
			d.Ack(false)
		}
	}()

	<-forever
}
