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
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type consumer struct {
	service OrdersService
}

func NewConsumer(service OrdersService) *consumer {
	return &consumer{service}
}

func (c *consumer) Listen(ctx context.Context, r *kafka.Reader, producer *broker.Producer) {
	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("[Orders Consumer] Error fetching Kafka message: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		msgCtx := broker.ExtractKafkaHeaders(ctx, msg.Headers)
		tr := otel.Tracer("kafka")
		spanCtx, span := tr.Start(msgCtx, fmt.Sprintf("Kafka - consume - %s", msg.Topic))

		o := &pb.Order{}
		if err := json.Unmarshal(msg.Value, o); err != nil {
			log.Printf("failed to unmarshal order: %v", err)
			_ = r.CommitMessages(ctx, msg)
			span.End()
			continue
		}

		switch msg.Topic {
		case broker.OrderCancelledEvent, broker.OrderFailedEvent, broker.OrderStockFailedEvent:
			o.Status = "cancelled"
			log.Printf("[SAGA] Marking order #%s as cancelled due to %s", o.ID, msg.Topic)
		case broker.OrderRefundedEvent:
			o.Status = "refunded"
			log.Printf("[SAGA] Marking order #%s as refunded", o.ID)
		case broker.OrderPaidEvent:
			o.Status = "paid"
		}

		_, err = c.service.UpdateOrder(spanCtx, o)
		if err != nil {
			log.Printf("failed to update order: %v", err)

			if retryErr := broker.HandleKafkaRetry(spanCtx, producer, msg.Topic, msg); retryErr != nil {
				log.Printf("Error handling Kafka retry: %v", retryErr)
			}
			_ = r.CommitMessages(ctx, msg)
			span.End()
			continue
		}

		span.AddEvent(fmt.Sprintf("order.%s", o.Status))
		span.End()

		log.Printf("[SAGA STATE] Order #%s transitioned to status: %s", o.ID, o.Status)
		_ = r.CommitMessages(ctx, msg)
	}
}
