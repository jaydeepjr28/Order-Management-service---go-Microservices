package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	"github.com/jaydeepjr28/omsv2-stock/gateway"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	serviceName    = "stock"
	grpcAddr       = common.EnvString("GRPC_ADDR", "localhost:2002")
	consulAddr     = common.EnvString("CONSUL_ADDR", "localhost:8500")
	kafkaBrokers   = common.EnvString("KAFKA_BROKERS", "localhost:9092")
	jaegerAddr     = common.EnvString("JAEGER_ADDR", "localhost:4318")
	advertisedAddr = common.EnvString("ADVERTISED_ADDR", "localhost:2002")
	dbURI          = common.EnvString("DB_URI", "postgres://postgres:password@localhost:5436/stock?sslmode=disable")
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

	brokersList := strings.Split(kafkaBrokers, ",")
	producer := broker.NewProducer(brokersList)
	defer producer.Close()

	stockTopics := []string{
		broker.OrderPaidEvent,
		broker.OrderCancelledEvent,
		broker.OrderFailedEvent,
	}
	reader := broker.NewConsumerGroup(brokersList, "stock-service-group", stockTopics)
	defer reader.Close()

	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	defer l.Close()

	store := NewStore(dbURI)
	svc := NewService(store)
	svcWithTelemetry := NewTelemetryMiddleware(svc)

	NewGRPCHandler(grpcServer, svcWithTelemetry)

	ordersGateway := gateway.New(registry)
	consumer := NewConsumer(svcWithTelemetry, ordersGateway)

	go consumer.Listen(ctx, reader, producer)

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
