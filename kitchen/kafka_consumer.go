package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/omsv2-kitchen/gateway"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type Consumer struct {
	gateway gateway.KitchenGateway
}

func NewConsumer(gateway gateway.KitchenGateway) *Consumer {
	return &Consumer{gateway}
}

func (c *Consumer) Listen(ctx context.Context, r *kafka.Reader, producer *broker.Producer) {
	log.Printf("[Kitchen Consumer] Kafka listening for events on group kitchen-service-group...")
	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("[Kitchen Consumer] Error fetching Kafka message: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		msgCtx := broker.ExtractKafkaHeaders(ctx, msg.Headers)
		tr := otel.Tracer("kafka")
		spanCtx, span := tr.Start(msgCtx, fmt.Sprintf("Kafka - consume - %s", msg.Topic))

		var o *pb.Order
		if err := json.Unmarshal(msg.Value, &o); err != nil {
			log.Printf("Error unmarshalling order: %v", err)
			_ = r.CommitMessages(ctx, msg)
			span.End()
			continue
		}

		switch msg.Topic {
		case broker.OrderCancelledEvent, broker.OrderFailedEvent:
			log.Printf("[SAGA] Kitchen: Received %s for order #%s. Meal preparation cancelled/aborted.", msg.Topic, o.ID)
			span.AddEvent("saga.kitchen.aborted")
			span.End()
			_ = r.CommitMessages(ctx, msg)

		case broker.OrderPaidEvent:
			if o.Status == "paid" {
				cookOrder()

				span.AddEvent(fmt.Sprintf("Order Cooked: %v", o))

				if err := c.gateway.UpdateOrder(spanCtx, &pb.Order{
					Status:     "ready",
					ID:         o.ID,
					CustomerID: o.CustomerID,
				}); err != nil {
					log.Printf("error updating the order %v: %v", o, err)

					if retryErr := broker.HandleKafkaRetry(spanCtx, producer, msg.Topic, msg); retryErr != nil {
						log.Printf("error handling Kafka retry: %v", retryErr)
					}
					_ = r.CommitMessages(ctx, msg)
					span.End()
					continue
				}
			}

			span.AddEvent(fmt.Sprintf("order.updated: %v", o))
			span.End()
			_ = r.CommitMessages(ctx, msg)

		default:
			_ = r.CommitMessages(ctx, msg)
			span.End()
		}
	}
}

func cookOrder() {
	log.Println("Cooking order...")
	time.Sleep(5 * time.Second)
	log.Println("Order cooked!")
}
