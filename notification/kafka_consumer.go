package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/jaydeepjr28/commons/broker"
	"github.com/segmentio/kafka-go"
)

type SMTPConfig struct {
	Host       string
	Port       string
	Email      string
	Password   string
	SenderName string
}

type Consumer struct {
	reader  *kafka.Reader
	smtpCfg SMTPConfig
}

func NewConsumer(brokers []string, smtpCfg SMTPConfig) *Consumer {
	topics := []string{
		broker.OrderCreatedEvent,
		broker.UserRegisteredEvent,
		broker.UserOTPRequestedEvent,
	}
	r := broker.NewConsumerGroup(brokers, "notification-service-group", topics)
	return &Consumer{reader: r, smtpCfg: smtpCfg}
}

func (c *Consumer) Listen(ctx context.Context) {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("[Notification Service] Error fetching Kafka message: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		switch msg.Topic {
		case broker.UserOTPRequestedEvent:
			var payload broker.UserOTPRequestedPayload
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Printf("[Notification Service] Error unmarshaling OTP event: %v", err)
				_ = c.reader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("==================================================================")
			log.Printf("  STRATUM AUTHENTICATION: REGISTRATION EMAIL VERIFICATION")
			log.Printf("  Recipient : %s (%s)", payload.Email, payload.Name)
			log.Printf("  Your Code : [%s]", payload.OTP)
			log.Printf("  Expires In: %d seconds (5 minutes)", payload.ExpiresInSeconds)
			log.Printf("==================================================================")

			subject := "Your STRATUM Verification Code"
			textBody := fmt.Sprintf("Hello %s,\n\nYour STRATUM verification code is: %s\nThis code will expire in %d seconds.\n\nIf you did not request this code, please ignore this email.", payload.Name, payload.OTP, payload.ExpiresInSeconds)
			htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #fafaf8; padding: 30px; margin: 0;">
  <div style="max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 12px; border: 1px solid #e7e5e4; padding: 32px; box-shadow: 0 4px 12px rgba(0,0,0,0.05);">
    <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 24px;">
      <span style="background: #0f4a36; color: #ffffff; font-weight: 800; padding: 6px 10px; border-radius: 6px; font-size: 14px;">S</span>
      <span style="font-size: 18px; font-weight: 700; color: #1c1917; letter-spacing: -0.5px;">STRATUM</span>
    </div>
    <h2 style="font-size: 20px; font-weight: 700; color: #1c1917; margin-top: 0;">Confirm Your Email</h2>
    <p style="font-size: 14px; color: #57534e; line-height: 1.5;">Hello <strong>%s</strong>,<br>Use the verification code below to complete your registration on the STRATUM platform.</p>
    <div style="background: #f5f5f4; border-radius: 8px; padding: 16px; text-align: center; margin: 24px 0;">
      <span style="font-size: 32px; font-weight: 800; letter-spacing: 6px; color: #0f4a36; font-family: monospace;">%s</span>
    </div>
    <p style="font-size: 12px; color: #78716c; line-height: 1.4;">This code is valid for <strong>%d seconds (5 minutes)</strong>. Do not share this code with anyone.</p>
    <hr style="border: none; border-top: 1px solid #e7e5e4; margin: 24px 0;">
    <p style="font-size: 11px; color: #a8a29e; text-align: center;">© STRATUM Commerce. Distributed Order & Fulfillment Platform.</p>
  </div>
</body>
</html>`, payload.Name, payload.OTP, payload.ExpiresInSeconds)

			c.sendEmail(payload.Email, subject, textBody, htmlBody)

		case broker.UserRegisteredEvent:
			var payload broker.UserRegisteredPayload
			recipient := "customer@stratum.com"
			name := "Valued Customer"

			if err := json.Unmarshal(msg.Value, &payload); err == nil && payload.Email != "" {
				recipient = payload.Email
				if payload.Name != "" {
					name = payload.Name
				}
			} else {
				var fallback map[string]interface{}
				if err := json.Unmarshal(msg.Value, &fallback); err == nil {
					if cid, ok := fallback["customerID"].(string); ok && cid != "" {
						recipient = cid
					}
				}
			}

			log.Printf("[Notification Service] Sending Welcome Email to %s (%s)", recipient, name)

			subject := "Welcome to STRATUM!"
			textBody := fmt.Sprintf("Welcome to STRATUM, %s!\n\nYour account has been registered and verified.\nExplore our curated provisions and real-time order tracking: http://localhost:5173", name)
			htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #fafaf8; padding: 30px; margin: 0;">
  <div style="max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 12px; border: 1px solid #e7e5e4; padding: 32px; box-shadow: 0 4px 12px rgba(0,0,0,0.05);">
    <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 24px;">
      <span style="background: #0f4a36; color: #ffffff; font-weight: 800; padding: 6px 10px; border-radius: 6px; font-size: 14px;">S</span>
      <span style="font-size: 18px; font-weight: 700; color: #1c1917; letter-spacing: -0.5px;">STRATUM</span>
    </div>
    <h2 style="font-size: 20px; font-weight: 700; color: #1c1917; margin-top: 0;">Welcome to STRATUM, %s!</h2>
    <p style="font-size: 14px; color: #57534e; line-height: 1.5;">Your account is active. You can now explore handcrafted culinary provisions, build your real-time cart, and track orders directly from kitchen to table.</p>
    <div style="text-align: center; margin: 28px 0;">
      <a href="http://localhost:5173/products" style="background: #0f4a36; color: #ffffff; text-decoration: none; padding: 12px 24px; border-radius: 8px; font-size: 14px; font-weight: 600; display: inline-block;">Browse Menu Catalog</a>
    </div>
    <hr style="border: none; border-top: 1px solid #e7e5e4; margin: 24px 0;">
    <p style="font-size: 11px; color: #a8a29e; text-align: center;">© STRATUM Commerce. Distributed Order & Fulfillment Platform.</p>
  </div>
</body>
</html>`, name)

			c.sendEmail(recipient, subject, textBody, htmlBody)

		case broker.OrderCreatedEvent:
			log.Printf("[Notification Service] Order Created event received: %s", string(msg.Value))

		default:
			log.Printf("[Notification Service] Event received on topic: %s", msg.Topic)
		}

		_ = c.reader.CommitMessages(ctx, msg)
	}
}

func (c *Consumer) sendEmail(to, subject, bodyText, bodyHTML string) {
	if c.smtpCfg.Email == "" || c.smtpCfg.Password == "" {
		log.Printf("--> [DEV STUB (No SMTP Credentials)] Dispatching to: %s", to)
		log.Printf("    Subject: %s", subject)
		log.Printf("    Message: %s", bodyText)
		return
	}

	fromHeader := fmt.Sprintf("%s <%s>", c.smtpCfg.SenderName, c.smtpCfg.Email)
	addr := fmt.Sprintf("%s:%s", c.smtpCfg.Host, c.smtpCfg.Port)
	auth := smtp.PlainAuth("", c.smtpCfg.Email, c.smtpCfg.Password, c.smtpCfg.Host)

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n\r\n", fromHeader, to, subject, boundary)
	msg += fmt.Sprintf("--%s\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s\r\n\r\n", boundary, bodyText)
	if bodyHTML != "" {
		msg += fmt.Sprintf("--%s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s\r\n\r\n", boundary, bodyHTML)
	}
	msg += fmt.Sprintf("--%s--\r\n", boundary)

	err := smtp.SendMail(addr, auth, c.smtpCfg.Email, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("[Notification Service] ❌ Failed to dispatch SMTP email to %s: %v", to, err)
		return
	}
	log.Printf("[Notification Service] ✅ Successfully sent email to %s via %s", to, c.smtpCfg.Host)
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
