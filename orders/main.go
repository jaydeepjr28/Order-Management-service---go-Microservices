package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	"github.com/jaydeepjr28/omsv2-orders/gateway"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	serviceName	= "orders"
	grpcAddr	= common.EnvString("GRPC_ADDR", "localhost:2000")
	consulAddr	= common.EnvString("CONSUL_ADDR", "localhost:8500")
	amqpUser	= common.EnvString("RABBITMQ_USER", "guest")
	amqpPass	= common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost	= common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort	= common.EnvString("RABBITMQ_PORT", "5672")
	jaegerAddr	= common.EnvString("JAEGER_ADDR", "localhost:4318")
	advertisedAddr	= common.EnvString("ADVERTISED_ADDR", "localhost:2000")
)

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	if err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr); err != nil {
		logger.Fatal("could set global tracer", zap.Error(err))
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

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logger.Error("Failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	ch, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	defer l.Close()

	gateway := gateway.NewGateway(registry)

	pgURI := common.EnvString("DB_URI", "postgres://postgres:password@localhost:5433/orders?sslmode=disable")
	store, err := NewStore(pgURI)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	svc := NewService(store, gateway)
	svcWithTelemetry := NewTelemetryMiddleware(svc)
	svcWithLogging := NewLoggingMiddleware(svcWithTelemetry)

	NewGRPCHandler(grpcServer, svcWithLogging, ch)

	consumer := NewConsumer(svcWithLogging)
	go consumer.Listen(ch)

	logger.Info("Starting gRPC server", zap.String("port", grpcAddr))

	go func() {
		if err := grpcServer.Serve(l); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	grpcServer.GracefulStop()
	logger.Info("Server exiting")
}
