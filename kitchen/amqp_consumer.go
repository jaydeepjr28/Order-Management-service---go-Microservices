package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/omsv2-kitchen/gateway"
	"go.opentelemetry.io/otel"
)

// Consumer listens for OrderPaidEvents to simulate food preparation,
// then updates the order status to 'ready' via the Orders gateway.
type Consumer struct {
	gateway gateway.KitchenGateway
}

func NewConsumer(gateway gateway.KitchenGateway) *Consumer {
	return &Consumer{gateway}
}

// Listen sets up the AMQP queue and starts processing incoming events.
func (c *Consumer) Listen(ch *amqp.Channel) {
	// Anonymous queue that auto-deletes on disconnect, routed to DLX on failure
	q, err := ch.QueueDeclare(
		"",   
		true, 
		false,
		true, 
		false,
		amqp.Table{"x-dead-letter-exchange": "dlx_main"},  
	)
	if err != nil {
		log.Fatal(err)
	}

	// Bind to the fanout exchange where OrderPaidEvents are published
	err = ch.QueueBind(
		q.Name,               
		"",                   
		broker.OrderPaidEvent,
		false,                
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			tr := otel.Tracer("amqp")
			// Start a new trace span for the cooking process
			_, messageSpan := tr.Start(context.Background(), fmt.Sprintf("AMQP - consume - %s", q.Name))

			var o *pb.Order
			if err := json.Unmarshal(d.Body, &o); err != nil {
				log.Printf("Error unmarshalling order: %v", err)
				d.Nack(false, false) // Bad data -> DLQ
				continue
			}

			// The kitchen only prepares orders that have been successfully paid
			if o.Status == "paid" {
				cookOrder()

				messageSpan.AddEvent(fmt.Sprintf("Order Cooked: %v", o))

				// Notify the Orders service that the food is ready for pickup/delivery
				if err := c.gateway.UpdateOrder(context.Background(), &pb.Order{
					Status:     "ready",
					ID:         o.ID,
					CustomerID: o.CustomerID,
				}); err != nil {
					log.Printf("error updating the order %v", o)

					// If the Orders service is down, retry with exponential backoff
					if err := broker.HandleRetry(ch, &d); err != nil {
						log.Printf("error handling the retry: %v", err.Error())
					}
					continue
				}
			}

			messageSpan.AddEvent(fmt.Sprintf("order.updated: %v", o))
			messageSpan.End()

			d.Ack(false) // Processing complete
		}
	}()

	log.Printf("AMQP Listening. To exit press CTRL+C")
	<-forever
}

// cookOrder is a mock function that simulates the time taken to prepare food.
func cookOrder() {
	log.Println("Cooking order...")
	time.Sleep(5 * time.Second)
	log.Println("Order cooked!")
}
