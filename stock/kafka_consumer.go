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
	"github.com/jaydeepjr28/omsv2-stock/gateway"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type Consumer struct {
	service StockService
	gateway gateway.OrdersGateway
}

func NewConsumer(service StockService, ordersGateway gateway.OrdersGateway) *Consumer {
	return &Consumer{service: service, gateway: ordersGateway}
}

func (c *Consumer) Listen(ctx context.Context, r *kafka.Reader, producer *broker.Producer) {
	log.Printf("[Stock Consumer] Kafka listening for events on group stock-service-group...")
	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("[Stock Consumer] Error fetching Kafka message: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		msgCtx := broker.ExtractKafkaHeaders(ctx, msg.Headers)
		tr := otel.Tracer("kafka")
		spanCtx, span := tr.Start(msgCtx, fmt.Sprintf("Kafka - consume - %s", msg.Topic))

		var o pb.Order
		if err := json.Unmarshal(msg.Value, &o); err != nil {
			log.Printf("failed to unmarshal order: %v", err)
			_ = r.CommitMessages(ctx, msg)
			span.End()
			continue
		}

		switch msg.Topic {
		case broker.OrderCancelledEvent, broker.OrderFailedEvent:
			log.Printf("[SAGA] Received %s for order #%s. Executing stock compensation...", msg.Topic, o.ID)
			if len(o.Items) == 0 {
				fullOrder, err := c.gateway.GetOrder(spanCtx, o.ID, o.CustomerID)
				if err == nil && fullOrder != nil {
					o.Items = fullOrder.Items
				}
			}
			if len(o.Items) > 0 {
				if err := c.service.IncrementItems(spanCtx, o.Items); err != nil {
					log.Printf("[SAGA COMPENSATION ERROR] Failed to increment stock for order %s: %v", o.ID, err)
				} else {
					log.Printf("[SAGA COMPENSATION SUCCESS] Restored stock inventory for order #%s", o.ID)
				}
			}
			span.AddEvent("saga.stock.compensated")
			span.End()
			_ = r.CommitMessages(ctx, msg)

		case broker.OrderPaidEvent:
			if o.Status != "paid" {
				_ = r.CommitMessages(ctx, msg)
				span.End()
				continue
			}

			if len(o.Items) == 0 {
				fullOrder, err := c.gateway.GetOrder(spanCtx, o.ID, o.CustomerID)
				if err != nil {
					log.Printf("failed to fetch order %s: %v", o.ID, err)
					if retryErr := broker.HandleKafkaRetry(spanCtx, producer, msg.Topic, msg); retryErr != nil {
						log.Printf("Error handling Kafka retry: %v", retryErr)
					}
					_ = r.CommitMessages(ctx, msg)
					span.End()
					continue
				}
				o.Items = fullOrder.Items
			}

			if err := c.service.DecrementItems(spanCtx, o.Items); err != nil {
				log.Printf("[SAGA] Failed to decrement stock for order %s: %v. Emitting failure events...", o.ID, err)
				marshalledOrder, _ := json.Marshal(&o)
				if producer != nil {
					_ = producer.Publish(spanCtx, broker.OrderStockFailedEvent, []byte(o.ID), marshalledOrder)
					_ = producer.Publish(spanCtx, broker.OrderFailedEvent, []byte(o.ID), marshalledOrder)
				}
				span.AddEvent("saga.stock.failed")
				span.End()
				_ = r.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Stock decremented for order %s", o.ID)
			span.AddEvent("stock.decremented")
			span.End()
			_ = r.CommitMessages(ctx, msg)

		default:
			_ = r.CommitMessages(ctx, msg)
			span.End()
		}
	}
}
