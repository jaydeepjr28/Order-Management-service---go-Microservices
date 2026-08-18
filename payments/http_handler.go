package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

type PaymentHTTPHandler struct {
	channel		*amqp.Channel
	webhookSecret	string
	store		PaymentStore
}

type razorpayWebhook struct {
	Event	string	`json:"event"`
	Payload	struct {
		PaymentLink struct {
			Entity struct {
				ReferenceID	string			`json:"reference_id"`
				Notes		map[string]string	`json:"notes"`
			} `json:"entity"`
		} `json:"payment_link"`
	}	`json:"payload"`
}

func NewPaymentHTTPHandler(channel *amqp.Channel, webhookSecret string, store PaymentStore) *PaymentHTTPHandler {
	return &PaymentHTTPHandler{channel: channel, webhookSecret: webhookSecret, store: store}
}

func (h *PaymentHTTPHandler) registerRoutes(router *gin.Engine) {
	router.POST("/webhook", h.handlePaymentWebhook)
}

func (h *PaymentHTTPHandler) handlePaymentWebhook(c *gin.Context) {
	if h.webhookSecret == "" {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 65536)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if !validWebhookSignature(body, c.GetHeader("X-Razorpay-Signature"), h.webhookSecret) {
		c.Status(http.StatusBadRequest)
		return
	}

	eventID := c.GetHeader("x-razorpay-event-id")
	if eventID != "" {
		has, err := h.store.HasEvent(c.Request.Context(), eventID)
		if err == nil && has {
			c.Status(http.StatusOK)
			return
		}
	}

	var event razorpayWebhook
	if err := json.Unmarshal(body, &event); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if event.Event != "payment_link.paid" {
		c.Status(http.StatusOK)
		return
	}

	orderID := event.Payload.PaymentLink.Entity.Notes["order_id"]
	customerID := event.Payload.PaymentLink.Entity.Notes["customer_id"]
	if orderID == "" || customerID == "" {
		log.Printf("Razorpay payment link webhook missing order or customer ID")
		c.Status(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := h.publishOrderPaid(ctx, &pb.Order{ID: orderID, CustomerID: customerID, Status: "paid"}); err != nil {
		log.Printf("failed to publish order.paid: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if eventID != "" {
		_ = h.store.SaveEvent(c.Request.Context(), eventID)
	}

	log.Printf("Razorpay payment received for order %s", orderID)
	c.Status(http.StatusOK)
}

func validWebhookSignature(body []byte, receivedSignature, secret string) bool {
	if receivedSignature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected, err := hex.DecodeString(receivedSignature)
	if err != nil {
		return false
	}
	return hmac.Equal(mac.Sum(nil), expected)
}

func (h *PaymentHTTPHandler) publishOrderPaid(ctx context.Context, order *pb.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal paid order: %w", err)
	}

	tr := otel.Tracer("amqp")
	ctx, span := tr.Start(ctx, fmt.Sprintf("AMQP - publish - %s", broker.OrderPaidEvent))
	defer span.End()

	return h.channel.PublishWithContext(ctx, broker.OrderPaidEvent, "", false, false, amqp.Publishing{
		ContentType:	"application/json",
		Body:		body,
		DeliveryMode:	amqp.Persistent,
		Headers:	broker.InjectAMQPHeaders(ctx),
	})
}
