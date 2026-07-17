package razorpay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "github.com/jaydeepjr28/commons/api"
)

const paymentLinksURL = "https://api.razorpay.com/v1/payment_links"

type Processor struct {
	keyID       string
	keySecret   string
	itemPrices  map[string]int64
	callbackURL string
	httpClient  *http.Client
}

type paymentLinkRequest struct {
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	ReferenceID    string            `json:"reference_id"`
	Description    string            `json:"description"`
	CallbackURL    string            `json:"callback_url"`
	CallbackMethod string            `json:"callback_method"`
	Notes          map[string]string `json:"notes"`
}

type paymentLinkResponse struct {
	ShortURL string `json:"short_url"`
}

func NewProcessor(keyID, keySecret, itemPricesJSON, gatewayHTTPAddr string) (*Processor, error) {
	if keyID == "" || keySecret == "" {
		return nil, fmt.Errorf("RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET are required")
	}

	prices := map[string]int64{}
	if err := json.Unmarshal([]byte(itemPricesJSON), &prices); err != nil {
		return nil, fmt.Errorf("RAZORPAY_ITEM_PRICES must be a JSON object of item IDs to paise amounts: %w", err)
	}
	if len(prices) == 0 {
		return nil, fmt.Errorf("RAZORPAY_ITEM_PRICES must define at least one item price")
	}
	for itemID, price := range prices {
		if itemID == "" || price < 100 {
			return nil, fmt.Errorf("invalid price for item %q: amount must be at least 100 paise", itemID)
		}
	}

	return &Processor{
		keyID:       keyID,
		keySecret:   keySecret,
		itemPrices:  prices,
		callbackURL: strings.TrimRight(gatewayHTTPAddr, "/") + "/success.html",
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (p *Processor) CreatePaymentLink(order *pb.Order) (string, error) {
	amount, description, err := p.orderTotal(order)
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(paymentLinkRequest{
		Amount:         amount,
		Currency:       "INR",
		ReferenceID:    order.ID,
		Description:    description,
		CallbackURL:    fmt.Sprintf("%s?customerID=%s&orderID=%s", p.callbackURL, order.CustomerID, order.ID),
		CallbackMethod: "get",
		Notes: map[string]string{
			"order_id":    order.ID,
			"customer_id": order.CustomerID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal Razorpay payment link: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, paymentLinksURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create Razorpay request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(p.keyID, p.keySecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call Razorpay payment links API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read Razorpay response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Razorpay payment link API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var result paymentLinkResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse Razorpay response: %w", err)
	}
	if result.ShortURL == "" {
		return "", fmt.Errorf("Razorpay response did not include short_url")
	}
	return result.ShortURL, nil
}

func (p *Processor) orderTotal(order *pb.Order) (int64, string, error) {
	var amount int64
	items := make([]string, 0, len(order.Items))
	for _, item := range order.Items {
		price, ok := p.itemPrices[item.ID]
		if !ok {
			return 0, "", fmt.Errorf("no Razorpay price configured for item %q", item.ID)
		}
		amount += price * int64(item.Quantity)
		items = append(items, fmt.Sprintf("%s x%d", item.Name, item.Quantity))
	}
	if amount < 100 {
		return 0, "", fmt.Errorf("order total must be at least 100 paise")
	}
	return amount, "Order: " + strings.Join(items, ", "), nil
}
