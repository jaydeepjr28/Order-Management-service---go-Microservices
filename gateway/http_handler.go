package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/auth"
	"github.com/jaydeepjr28/commons/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/jaydeepjr28/omsv2-gateway/gateway"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// handler holds all the dependencies required to process HTTP requests.
// It acts as the API gateway, routing requests to the appropriate downstream gRPC services
// (like Orders or Cart) or handling them locally (like User Auth using MongoDB).
type handler struct {
	ordersGateway gateway.OrdersGateway
	store         *store
	cartGateway   gateway.CartGateway
	publicURL     string
	channel       *amqp.Channel
}

// NewHandler creates a new HTTP handler with all required microservice clients and local store.
func NewHandler(orders gateway.OrdersGateway, store *store, cart gateway.CartGateway, publicURL string, channel *amqp.Channel) *handler {
	return &handler{ordersGateway: orders, store: store, cartGateway: cart, publicURL: publicURL, channel: channel}
}

// writeGRPCError translates gRPC status codes from downstream services into standard HTTP status codes
// so the frontend client receives a standardized REST-like error response.
func writeGRPCError(w http.ResponseWriter, err error) {
	st := status.Convert(err)
	switch st.Code() {
	case codes.InvalidArgument, codes.FailedPrecondition:
		common.WriteError(w, http.StatusBadRequest, st.Message())
	case codes.NotFound:
		common.WriteError(w, http.StatusNotFound, st.Message())
	case codes.Unauthenticated:
		common.WriteError(w, http.StatusUnauthorized, st.Message())
	default:
		common.WriteError(w, http.StatusInternalServerError, st.Message())
	}
}

// authMiddleware is a Gin middleware that intercepts incoming requests,
// extracts the Bearer JWT token from the Authorization header, validates it,
// and injects the authenticated customerID into the request context.
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		// Validate token against our symmetric secret
		customerID, err := auth.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Attach the customerID to the Gin context so downstream handlers can access it securely
		c.Set("customerID", customerID)
		c.Next()
	}
}

func (h *handler) registerRoutes(router *gin.Engine) {
	router.POST("/api/users/register", h.handleRegister)
	router.POST("/api/users/login", h.handleLogin)

	api := router.Group("/api")
	api.Use(authMiddleware())
	{
		api.POST("/customers/:customerID/orders", h.handleCreateOrder)
		api.GET("/customers/:customerID/orders", h.handleGetCustomerOrders)
		api.GET("/customers/:customerID/orders/:orderID", h.handleGetOrder)
		api.GET("/customers/:customerID/orders/:orderID/status", h.handleGetOrderStatus)
		api.GET("/inventory", h.handleGetInventory)

		api.GET("/cart", h.handleGetCart)
		api.POST("/cart/items", h.handleAddToCart)
		api.DELETE("/cart/items/:productID", h.handleRemoveFromCart)
		api.DELETE("/cart", h.handleClearCart)
	}

}


// handleRegister creates a new user account in MongoDB and issues an initial JWT token.
func (h *handler) handleRegister(c *gin.Context) {
	var req pb.RegisterRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	u := &UserModel{
		Email:    req.Email,
		Password: req.Password, // Note: Store handles hashing the password before saving
		Name:     req.Name,
	}

	if err := h.store.CreateUser(c.Request.Context(), u); err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	// Issue a 24-hour token for the new user
	token, err := auth.GenerateToken(u.Email, 24*time.Hour)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	res := &pb.AuthResponse{
		Token:      token,
		CustomerID: u.Email,
	}

	// Publish UserRegisteredEvent
	if h.channel != nil {
		tr := otel.Tracer("amqp")
		amqpCtx, span := tr.Start(c.Request.Context(), fmt.Sprintf("AMQP - publish - %s", broker.UserRegisteredEvent))
		headers := broker.InjectAMQPHeaders(amqpCtx)
		
		body, _ := json.Marshal(res) // Send AuthResponse as event payload
		
		h.channel.PublishWithContext(amqpCtx, broker.UserRegisteredEvent, "", false, false, amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
		})
		span.End()
	}

	common.WriteJSON(c.Writer, http.StatusOK, res)
}

// handleLogin verifies user credentials against the MongoDB database and returns a JWT token.
func (h *handler) handleLogin(c *gin.Context) {
	var req pb.LoginRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	u, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		common.WriteError(c.Writer, http.StatusNotFound, "user not found")
		return
	}

	// Compare the provided plain text password with the bcrypt hash stored in the DB
	if err := h.store.VerifyPassword(u.Password, req.Password); err != nil {
		common.WriteError(c.Writer, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(u.Email, 24*time.Hour)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	res := &pb.AuthResponse{
		Token:      token,
		CustomerID: u.Email,
	}
	common.WriteJSON(c.Writer, http.StatusOK, res)
}


// handleGetCart retrieves the current state of a user's shopping cart from the downstream Cart service.
func (h *handler) handleGetCart(c *gin.Context) {
	// Extract the securely verified customerID from the context (injected by authMiddleware)
	customerID := c.GetString("customerID")
	res, err := h.cartGateway.GetCart(c.Request.Context(), &pb.GetCartRequest{CustomerID: customerID})
	if err != nil {
		writeGRPCError(c.Writer, err)
		return
	}
	common.WriteJSON(c.Writer, http.StatusOK, res)
}

func (h *handler) handleAddToCart(c *gin.Context) {
	customerID := c.GetString("customerID")
	var item pb.CartItem
	if err := common.ReadJSON(c.Request, &item); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.cartGateway.AddItem(c.Request.Context(), &pb.AddItemRequest{
		CustomerID: customerID,
		Item:       &item,
	})
	if err != nil {
		writeGRPCError(c.Writer, err)
		return
	}
	common.WriteJSON(c.Writer, http.StatusOK, res)
}

func (h *handler) handleRemoveFromCart(c *gin.Context) {
	customerID := c.GetString("customerID")
	productID := c.Param("productID")

	res, err := h.cartGateway.RemoveItem(c.Request.Context(), &pb.RemoveItemRequest{
		CustomerID: customerID,
		ProductID:  productID,
	})
	if err != nil {
		writeGRPCError(c.Writer, err)
		return
	}
	common.WriteJSON(c.Writer, http.StatusOK, res)
}

func (h *handler) handleClearCart(c *gin.Context) {
	customerID := c.GetString("customerID")
	res, err := h.cartGateway.ClearCart(c.Request.Context(), &pb.ClearCartRequest{CustomerID: customerID})
	if err != nil {
		writeGRPCError(c.Writer, err)
		return
	}
	common.WriteJSON(c.Writer, http.StatusOK, res)
}

// handleGetOrder fetches the details of a specific order.
func (h *handler) handleGetOrder(c *gin.Context) {
	customerID := c.Param("customerID")
	orderID := c.Param("orderID")

	// Security Check: Users can only request their own orders
	if customerID != c.GetString("customerID") {
		common.WriteError(c.Writer, http.StatusForbidden, "not allowed to access this order")
		return
	}

	// Start a trace span for distributed tracing through the gRPC call
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	o, err := h.ordersGateway.GetOrder(ctx, orderID, customerID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	common.WriteJSON(c.Writer, http.StatusOK, o)
}

// handleGetCustomerOrders fetches all orders for a specific customer.
func (h *handler) handleGetCustomerOrders(c *gin.Context) {
	customerID := c.Param("customerID")

	if customerID != c.GetString("customerID") {
		common.WriteError(c.Writer, http.StatusForbidden, "not allowed to access these orders")
		return
	}

	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	res, err := h.ordersGateway.GetCustomerOrders(ctx, customerID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	common.WriteJSON(c.Writer, http.StatusOK, res.Orders)
}

func (h *handler) handleGetOrderStatus(c *gin.Context) {
	customerID := c.Param("customerID")
	orderID := c.Param("orderID")

	if customerID != c.GetString("customerID") {
		common.WriteError(c.Writer, http.StatusForbidden, "not allowed to access this order")
		return
	}

	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	o, err := h.ordersGateway.GetOrder(ctx, orderID, customerID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	common.WriteJSON(c.Writer, http.StatusOK, gin.H{"status": o.Status})
}

func (h *handler) handleGetInventory(c *gin.Context) {
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	items, err := h.ordersGateway.GetItems(ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJSON(c.Writer, http.StatusOK, items)
}


// handleCreateOrder processes a checkout request.
// It proxies the request to the downstream Orders service via gRPC.
func (h *handler) handleCreateOrder(c *gin.Context) {
	customerID := c.Param("customerID")

	// Ensure the user is only submitting orders for their own account
	if customerID != c.GetString("customerID") {
		common.WriteError(c.Writer, http.StatusForbidden, "not allowed to access this order")
		return
	}

	var items []*pb.ItemsWithQuantity
	if err := common.ReadJSON(c.Request, &items); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	// Ensure the order has valid items and quantities before submitting to backend
	if err := validateItems(items); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	// Call the remote Order service over gRPC. 
	// This service will internally check stock, reserve it, and return a payment link.
	o, err := h.ordersGateway.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: customerID,
		Items:      items,
	})
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	// Return the created order along with the UI redirect URL for success
	res := &CreateOrderRequest{
		Order:         o,
		RedirectToURL: fmt.Sprintf("%s/success.html?customerID=%s&orderID=%s", h.publicURL, o.CustomerID, o.ID),
	}

	common.WriteJSON(c.Writer, http.StatusOK, res)
}

func validateItems(items []*pb.ItemsWithQuantity) error {
	if len(items) == 0 {
		return common.ErrNoItems
	}

	for _, i := range items {
		if i.ID == "" {
			return errors.New("item ID is required")
		}
		if i.Quantity <= 0 {
			return errors.New("items must have a valid quantity")
		}
	}

	return nil
}
