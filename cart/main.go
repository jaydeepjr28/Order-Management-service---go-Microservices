package main

import (
	"context"
	"log"
	"net"
	"time"

	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
)

var (
	serviceName	= "cart"
	grpcAddr	= common.EnvString("GRPC_ADDR", "localhost:2005")
	consulAddr	= common.EnvString("CONSUL_ADDR", "localhost:8500")
	jaegerAddr	= common.EnvString("JAEGER_ADDR", "localhost:4318")
	redisAddr	= common.EnvString("REDIS_ADDR", "localhost:6379")
	advertisedAddr	= common.EnvString("ADVERTISED_ADDR", "localhost:2005")
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

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Printf("failed to health check: %v", err)
			}
			time.Sleep(time.Second * 1)
		}
	}()

	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()

	store := NewRedisStore(redisAddr)
	NewGRPCHandler(grpcServer, store)

	log.Printf("Starting gRPC server at %s", grpcAddr)
	if err := grpcServer.Serve(l); err != nil {
		log.Fatal(err)
	}
}
