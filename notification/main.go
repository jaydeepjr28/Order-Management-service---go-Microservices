package main

import (
	"context"
	"log"
	"strings"

	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	_ "github.com/joho/godotenv/autoload"
)

var (
	serviceName    = "notification"
	consulAddr     = common.EnvString("CONSUL_ADDR", "localhost:8500")
	jaegerAddr     = common.EnvString("JAEGER_ADDR", "localhost:4318")
	kafkaBrokers   = common.EnvString("KAFKA_BROKERS", "localhost:9092")
	advertisedAddr = common.EnvString("ADVERTISED_ADDR", "localhost:2006")

	// Gmail / SMTP Configuration
	smtpHost       = common.EnvString("SMTP_HOST", "smtp.gmail.com")
	smtpPort       = common.EnvString("SMTP_PORT", "587")
	smtpEmail      = common.EnvString("SMTP_EMAIL", "")
	smtpPassword   = common.EnvString("SMTP_PASSWORD", "")
	smtpSenderName = common.EnvString("SMTP_SENDER_NAME", "STRATUM")
)

func main() {

	if err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr); err != nil {
		log.Printf("failed to set global tracer: %v", err)
	}

	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, advertisedAddr); err != nil {
		panic(err)
	}
	defer registry.Deregister(ctx, instanceID, serviceName)

	brokersList := strings.Split(kafkaBrokers, ",")
	smtpCfg := SMTPConfig{
		Host:       smtpHost,
		Port:       smtpPort,
		Email:      smtpEmail,
		Password:   smtpPassword,
		SenderName: smtpSenderName,
	}
	consumer := NewConsumer(brokersList, smtpCfg)
	defer consumer.Close()

	if smtpEmail != "" && smtpPassword != "" {
		log.Printf("Notification service configured with Gmail SMTP (%s) as %s", smtpHost, smtpEmail)
	} else {
		log.Println("Notification service running in Dev Sandbox mode (SMTP disabled, emails logged to terminal)")
	}

	log.Println("Notification service started, listening for Kafka events...")

	go consumer.Listen(ctx)

	select {}
}
