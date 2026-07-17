package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	_ "github.com/joho/godotenv/autoload"

	"github.com/jaydeepjr28/omsv2-payments/gateway"
	"github.com/jaydeepjr28/omsv2-payments/processor/razorpay"
	"google.golang.org/grpc"
)

var (
	serviceName           = "payment"
	amqpUser              = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass              = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost              = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort              = common.EnvString("RABBITMQ_PORT", "5672")
	grpcAddr              = common.EnvString("GRPC_ADDRESS", "localhost:2001")
	consulAddr            = common.EnvString("CONSUL_ADDR", "localhost:8500")
	httpAddr              = common.EnvString("HTTP_ADDR", "localhost:8081")
	razorpayKeyID         = common.EnvString("RAZORPAY_KEY_ID", "")
	razorpayKeySecret     = common.EnvString("RAZORPAY_KEY_SECRET", "")
	razorpayWebhookSecret = common.EnvString("RAZORPAY_WEBHOOK_SECRET", "")
	razorpayItemPrices    = common.EnvString("RAZORPAY_ITEM_PRICES", "")
	gatewayHTTPAddr       = common.EnvString("GATEWAY_HTTP_ADDRESS", "http://localhost:8080")
	jaegerAddr            = common.EnvString("JAEGER_ADDR", "localhost:4318")
)

func main() {
	if err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr); err != nil {
		log.Fatal("failed to set global tracer")
	}

	// Register consul
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	ctx := context.Background()
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Fatalf("failed to health check %v", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	// Broker connection
	ch, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	razorpayProcessor, err := razorpay.NewProcessor(razorpayKeyID, razorpayKeySecret, razorpayItemPrices, gatewayHTTPAddr)
	if err != nil {
		log.Fatalf("invalid Razorpay configuration: %v", err)
	}
	gateway := gateway.NewGateway(registry)
	svc := NewService(razorpayProcessor, gateway)
	svcWithTelemetry := NewTelemetryMiddleware(svc)

	amqpConsumer := NewConsumer(svcWithTelemetry)
	go amqpConsumer.Listen(ch)

	// http server
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	httpServer := NewPaymentHTTPHandler(ch, razorpayWebhookSecret)
	httpServer.registerRoutes(router)

	go func() {
		log.Printf("Starting HTTP server at %s", httpAddr)
		if err := router.Run(httpAddr); err != nil {
			log.Fatal("failed to start http server")
		}
	}()

	// gRPC server
	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()

	log.Println("GRPC Server Started at ", grpcAddr)
	if err := grpcServer.Serve(l); err != nil {
		log.Fatal(err.Error())
	}
}
