package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/omsv2-stock/gateway"
	"go.opentelemetry.io/otel"
)

type Consumer struct {
	service	StockService
	gateway	gateway.OrdersGateway
}

func NewConsumer(service StockService, ordersGateway gateway.OrdersGateway) *Consumer {
	return &Consumer{service: service, gateway: ordersGateway}
}

func (c *Consumer) Listen(ch *amqp.Channel) {

	q, err := ch.QueueDeclare("", true, false, true, false, amqp.Table{"x-dead-letter-exchange": "dlx_main"})
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(q.Name, "", broker.OrderPaidEvent, false, nil)
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
			ctx := broker.ExtractAMQPHeader(context.Background(), d.Headers)

			tr := otel.Tracer("amqp")
			_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - consume - %s", q.Name))

			var o pb.Order
			if err := json.Unmarshal(d.Body, &o); err != nil {
				log.Printf("failed to unmarshal order: %v", err)
				d.Nack(false, false)
				messageSpan.End()
				continue
			}

			if o.Status != "paid" {
				d.Ack(false)
				messageSpan.End()
				continue
			}

			if len(o.Items) == 0 {
				fullOrder, err := c.gateway.GetOrder(ctx, o.ID, o.CustomerID)
				if err != nil {
					log.Printf("failed to fetch order %s: %v", o.ID, err)
					if err := broker.HandleRetry(ch, &d); err != nil {
						log.Printf("error handling retry: %v", err)
					}
					messageSpan.End()
					continue
				}
				o.Items = fullOrder.Items
			}

			if err := c.service.DecrementItems(ctx, o.Items); err != nil {
				log.Printf("failed to decrement stock for order %s: %v", o.ID, err)
				if err := broker.HandleRetry(ch, &d); err != nil {
					log.Printf("error handling retry: %v", err)
				}
				messageSpan.End()
				continue
			}

			log.Printf("Stock decremented for order %s", o.ID)
			messageSpan.AddEvent("stock.decremented")
			messageSpan.End()
			d.Ack(false)
		}
	}()

	log.Printf("AMQP Listening. To exit press CTRL+C")
	<-forever
}
