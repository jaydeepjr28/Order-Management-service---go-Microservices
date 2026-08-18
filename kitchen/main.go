package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	"github.com/jaydeepjr28/omsv2-kitchen/gateway"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	serviceName = "kitchen"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2003")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	amqpUser    = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass    = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost    = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort       = common.EnvString("RABBITMQ_PORT", "5672")
	jaegerAddr     = common.EnvString("JAEGER_ADDR", "localhost:4318")
	advertisedAddr = common.EnvString("ADVERTISED_ADDR", "localhost:2003")
)

func main() {
	// Initialize structured logging using Zap
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	if err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr); err != nil {
		logger.Fatal("could set global tracer", zap.Error(err))
	}

	// Register this kitchen service with Consul for service discovery
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, advertisedAddr); err != nil {
		panic(err)
	}

	// Periodically ping Consul to stay registered as healthy
	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logger.Error("Failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	// Connect to RabbitMQ to listen for OrderPaidEvents
	ch, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	// Note: Kitchen doesn't actually expose any gRPC endpoints currently,
	// it only acts as a sink for AMQP events. The grpcServer is mostly boilerplate here.
	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	defer l.Close()

	// Initialize the gRPC gateway to communicate with downstream services (Orders)
	gateway := gateway.New(registry)

	// Start consuming AMQP messages (OrderPaidEvent)
	consumer := NewConsumer(gateway)
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
