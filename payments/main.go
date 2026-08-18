package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	_ "github.com/joho/godotenv/autoload"

	"github.com/jaydeepjr28/omsv2-payments/gateway"
	"github.com/jaydeepjr28/omsv2-payments/processor"
	"github.com/jaydeepjr28/omsv2-payments/processor/inmem"
	"github.com/jaydeepjr28/omsv2-payments/processor/razorpay"
	"google.golang.org/grpc"
)

var (
	serviceName		= "payment"
	amqpUser		= common.EnvString("RABBITMQ_USER", "guest")
	amqpPass		= common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost		= common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort		= common.EnvString("RABBITMQ_PORT", "5672")
	grpcAddr		= common.EnvString("GRPC_ADDRESS", "localhost:2001")
	consulAddr		= common.EnvString("CONSUL_ADDR", "localhost:8500")
	httpAddr		= common.EnvString("HTTP_ADDR", "localhost:8081")
	razorpayKeyID		= common.EnvString("RAZORPAY_KEY_ID", "")
	razorpayKeySecret	= common.EnvString("RAZORPAY_KEY_SECRET", "")
	razorpayWebhookSecret	= common.EnvString("RAZORPAY_WEBHOOK_SECRET", "")
	razorpayItemPrices	= common.EnvString("RAZORPAY_ITEM_PRICES", "")
	gatewayHTTPAddr		= common.EnvString("GATEWAY_HTTP_ADDRESS", "http://localhost:8080")
	jaegerAddr		= common.EnvString("JAEGER_ADDR", "localhost:4318")
	advertisedAddr		= common.EnvString("ADVERTISED_ADDR", "localhost:2001")
)

func main() {

	if err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr); err != nil {
		log.Fatal("failed to set global tracer")
	}

	pgURI := common.EnvString("DB_URI", "postgres://postgres:password@localhost:5434/payments?sslmode=disable")
	paymentStore, err := NewPaymentStore(pgURI)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	ctx := context.Background()
	if err := registry.Register(ctx, instanceID, serviceName, advertisedAddr); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Printf("failed to health check: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	ch, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	var paymentProcessor processor.PaymentProcessor
	if razorpayKeyID == "dummy_key" || razorpayKeyID == "" {
		paymentProcessor = inmem.NewInmem()
	} else {
		var err error
		paymentProcessor, err = razorpay.NewProcessor(razorpayKeyID, razorpayKeySecret, razorpayItemPrices, gatewayHTTPAddr)
		if err != nil {
			log.Fatalf("invalid Razorpay configuration: %v", err)
		}
	}

	gateway := gateway.NewGateway(registry)
	svc := NewService(paymentProcessor, gateway)
	svcWithTelemetry := NewTelemetryMiddleware(svc)

	amqpConsumer := NewConsumer(svcWithTelemetry)
	go amqpConsumer.Listen(ch)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	httpServer := NewPaymentHTTPHandler(ch, razorpayWebhookSecret, paymentStore)
	httpServer.registerRoutes(router)

	srv := &http.Server{
		Addr:		httpAddr,
		Handler:	router,
	}

	go func() {
		log.Printf("Starting HTTP server at %s", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start http server: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()

	log.Println("GRPC Server Started at ", grpcAddr)
	go func() {
		if err := grpcServer.Serve(l); err != nil {
			log.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("HTTP Server forced to shutdown: %v", err)
	}

	log.Println("Servers exited properly")
}
