package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/jaydeepjr28/commons"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/commons/discovery"
	"github.com/jaydeepjr28/commons/discovery/consul"
	"github.com/jaydeepjr28/omsv2-gateway/gateway"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

var (
	serviceName     = "gateway"
	httpAddr        = common.EnvString("HTTP_ADDR", ":8080")
	gatewayHTTPAddr = common.EnvString("GATEWAY_HTTP_ADDRESS", "http://localhost:8080")
	consulAddr      = common.EnvString("CONSUL_ADDR", "localhost:8500")
	jaegerAddr      = common.EnvString("JAEGER_ADDR", "localhost:4318")
	advertisedAddr  = common.EnvString("ADVERTISED_ADDR", "localhost:8080")
	kafkaBrokers       = common.EnvString("KAFKA_BROKERS", "localhost:9092")
	redisAddr          = common.EnvString("REDIS_ADDR", "localhost:6379")
	googleClientID     = common.EnvString("GOOGLE_CLIENT_ID", "")
	googleClientSecret = common.EnvString("GOOGLE_CLIENT_SECRET", "")
	googleRedirectURL  = common.EnvString("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback")
	frontendURL        = common.EnvString("FRONTEND_URL", "http://localhost:5173")
	adminEmail         = common.EnvString("ADMIN_EMAIL", "admin@stratum.com")
)

func main() {

	err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr)
	if err != nil {
		log.Fatal("failed to set global tracer")
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
				log.Printf("failed to health check: %v", err)
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), rateLimitMiddleware())

	ordersGateway := gateway.NewGRPCGateway(registry)

	pgURI := common.EnvString("DB_URI", "postgres://postgres:password@localhost:5435/gateway?sslmode=disable")
	store, err := NewStore(pgURI)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	brokersList := strings.Split(kafkaBrokers, ",")
	_ = broker.EnsureTopics(brokersList,
		broker.UserOTPRequestedEvent,
		broker.UserRegisteredEvent,
		broker.OrderCreatedEvent,
		broker.OrderPaidEvent,
		broker.OrderFailedEvent,
		broker.OrderCancelledEvent,
		broker.OrderRefundedEvent,
		broker.OrderStockFailedEvent,
		broker.DLQTopic,
	)

	producer := broker.NewProducer(brokersList)
	defer producer.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	oauthCfg := GoogleOAuthConfig{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		FrontendURL:  frontendURL,
	}

	handler := NewHandler(ordersGateway, store, ordersGateway, gatewayHTTPAddr, producer, redisClient, oauthCfg, adminEmail)
	handler.registerRoutes(router)

	srv := &http.Server{
		Addr:		httpAddr,
		Handler:	router,
	}

	go func() {
		log.Printf("Starting HTTP server at %s", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start http server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func rateLimitMiddleware() gin.HandlerFunc {
	limiter := rate.NewLimiter(5, 10)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}
