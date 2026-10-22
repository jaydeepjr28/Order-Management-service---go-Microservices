package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type consumer struct {
	service PaymentsService
}

func NewConsumer(service PaymentsService) *consumer {
	return &consumer{service}
}

func (c *consumer) Listen(ctx context.Context, r *kafka.Reader, producer *broker.Producer) {
	log.Printf("[Payments Consumer] Kafka listening for events on group payments-service-group...")
	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("[Payments Consumer] Error fetching Kafka message: %v", err)
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
			log.Printf("[SAGA] Received %s for order #%s. Executing payment compensation (refund)...", msg.Topic, o.ID)
			if err := c.service.RefundPayment(spanCtx, o); err != nil {
				log.Printf("[SAGA COMPENSATION ERROR] Failed to process refund for order %s: %v", o.ID, err)
			} else {
				log.Printf("[SAGA COMPENSATION SUCCESS] Issued payment refund for order #%s", o.ID)
				marshalledOrder, _ := json.Marshal(o)
				if producer != nil {
					_ = producer.Publish(spanCtx, broker.OrderRefundedEvent, []byte(o.ID), marshalledOrder)
				}
			}
			span.AddEvent("saga.payment.refunded")
			span.End()
			_ = r.CommitMessages(ctx, msg)

		case broker.OrderCreatedEvent:
			paymentLink, err := c.service.CreatePayment(spanCtx, o)
			if err != nil {
				log.Printf("failed to create payment: %v", err)

				if retryErr := broker.HandleKafkaRetry(spanCtx, producer, msg.Topic, msg); retryErr != nil {
					log.Printf("Error handling Kafka retry: %v", retryErr)
				}

				var retryCount int64 = 0
				for _, h := range msg.Headers {
					if h.Key == broker.RetryCountHeader {
						retryCount, _ = strconv.ParseInt(string(h.Value), 10, 64)
						break
					}
				}

				if retryCount >= broker.KafkaMaxRetryCount {
					log.Printf("[SAGA] Payment retries exhausted for order #%s. Emitting order.failed to trigger distributed compensation...", o.ID)
					marshalledOrder, _ := json.Marshal(o)
					if producer != nil {
						_ = producer.Publish(spanCtx, broker.OrderFailedEvent, []byte(o.ID), marshalledOrder)
					}
				}

				_ = r.CommitMessages(ctx, msg)
				span.End()
				continue
			}

			span.AddEvent(fmt.Sprintf("payment.created: %s", paymentLink))
			span.End()

			log.Printf("Payment link created %s for order %s", paymentLink, o.ID)
			_ = r.CommitMessages(ctx, msg)

		default:
			_ = r.CommitMessages(ctx, msg)
			span.End()
		}
	}
}
