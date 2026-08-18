package main

import (
	"context"
	"log"

	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	_ "github.com/joho/godotenv/autoload"
)

var (
	serviceName	= "notification"
	consulAddr	= common.EnvString("CONSUL_ADDR", "localhost:8500")
	jaegerAddr	= common.EnvString("JAEGER_ADDR", "localhost:4318")
	rabbitmqAddr	= common.EnvString("RABBITMQ_HOST", "localhost:5672")
	advertisedAddr	= common.EnvString("ADVERTISED_ADDR", "localhost:2006")
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

	consumer, err := NewConsumer(rabbitmqAddr)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	log.Println("Notification service started, listening for events...")

	go consumer.Listen()

	select {}
}
