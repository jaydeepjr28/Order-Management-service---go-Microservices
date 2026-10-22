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
)

type PaymentHTTPHandler struct {
	producer      *broker.Producer
	webhookSecret string
	keyID         string
	keySecret     string
	store         PaymentStore
	httpClient    *http.Client
}

type razorpayWebhook struct {
	Event   string `json:"event"`
	Payload struct {
		PaymentLink struct {
			Entity struct {
				ReferenceID string            `json:"reference_id"`
				Notes       map[string]string `json:"notes"`
			} `json:"entity"`
		} `json:"payment_link"`
	} `json:"payload"`
}

type verifyPaymentRequest struct {
	OrderID    string `json:"orderID"`
	CustomerID string `json:"customerID"`
}

func NewPaymentHTTPHandler(producer *broker.Producer, webhookSecret, keyID, keySecret string, store PaymentStore) *PaymentHTTPHandler {
	return &PaymentHTTPHandler{
		producer:      producer,
		webhookSecret: webhookSecret,
		keyID:         keyID,
		keySecret:     keySecret,
		store:         store,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *PaymentHTTPHandler) registerRoutes(router *gin.Engine) {
	router.POST("/webhook", h.handlePaymentWebhook)
	router.POST("/verify-payment", h.handleVerifyPayment)
}

func (h *PaymentHTTPHandler) handleVerifyPayment(c *gin.Context) {
	var req verifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OrderID == "" || req.CustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderID and customerID are required"})
		return
	}

	// In-memory / Mock offline mode
	if h.keyID == "" || h.keySecret == "" || h.keyID == "dummy_key" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err := h.publishOrderPaid(ctx, &pb.Order{ID: req.OrderID, CustomerID: req.CustomerID, Status: "paid"}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "paid", "verified": true})
		return
	}

	// Query Razorpay API for recent payment links to match notes.order_id
	url := "https://api.razorpay.com/v1/payment_links?count=25"
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpReq.SetBasicAuth(h.keyID, h.keySecret)

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to contact Razorpay: %v", err)})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var plinksResp struct {
		PaymentLinks []struct {
			ID     string            `json:"id"`
			Status string            `json:"status"`
			Notes  map[string]string `json:"notes"`
		} `json:"payment_links"`
	}

	if err := json.Unmarshal(body, &plinksResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse Razorpay response"})
		return
	}

	for _, plink := range plinksResp.PaymentLinks {
		if plink.Notes["order_id"] == req.OrderID && plink.Notes["customer_id"] == req.CustomerID {
			if plink.Status == "paid" {
				ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
				defer cancel()
				if err := h.publishOrderPaid(ctx, &pb.Order{ID: req.OrderID, CustomerID: req.CustomerID, Status: "paid"}); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "paid", "verified": true, "paymentLinkId": plink.ID})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": plink.Status, "verified": false, "message": "Payment has not been completed on Razorpay yet."})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "No matching payment link found in Razorpay for this order"})
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

	if h.producer != nil {
		return h.producer.Publish(ctx, broker.OrderPaidEvent, []byte(order.ID), body)
	}
	return nil
}
