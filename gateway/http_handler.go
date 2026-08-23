package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/auth"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/omsv2-gateway/gateway"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	ordersGateway	gateway.OrdersGateway
	store		*store
	cartGateway	gateway.CartGateway
	publicURL	string
	channel		*amqp.Channel
}

func NewHandler(orders gateway.OrdersGateway, store *store, cart gateway.CartGateway, publicURL string, channel *amqp.Channel) *handler {
	return &handler{ordersGateway: orders, store: store, cartGateway: cart, publicURL: publicURL, channel: channel}
}

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

		customerID, err := auth.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("customerID", customerID)
		c.Next()
	}
}

func (h *handler) registerRoutes(router *gin.Engine) {
	router.POST("/api/users/register", h.handleRegister)
	router.POST("/api/users/login", h.handleLogin)
	router.POST("/api/users/refresh", h.handleRefreshToken)

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

func (h *handler) handleRegister(c *gin.Context) {
	var req pb.RegisterRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	u := &UserModel{
		Email:		req.Email,
		Password:	req.Password,
		Name:		req.Name,
	}

	if err := h.store.CreateUser(c.Request.Context(), u); err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	accessToken, refreshToken, err := auth.GenerateAuthTokens(u.Email)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	res := &pb.AuthResponse{
		AccessToken:	accessToken,
		RefreshToken:	refreshToken,
		CustomerID:	u.Email,
	}

	if h.channel != nil {
		tr := otel.Tracer("amqp")
		amqpCtx, span := tr.Start(c.Request.Context(), fmt.Sprintf("AMQP - publish - %s", broker.UserRegisteredEvent))
		headers := broker.InjectAMQPHeaders(amqpCtx)

		body, _ := json.Marshal(res)

		h.channel.PublishWithContext(amqpCtx, broker.UserRegisteredEvent, "", false, false, amqp.Publishing{
			ContentType:	"application/json",
			Body:		body,
			DeliveryMode:	amqp.Persistent,
			Headers:	headers,
		})
		span.End()
	}

	common.WriteJSON(c.Writer, http.StatusOK, res)
}

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

	if err := h.store.VerifyPassword(u.Password, req.Password); err != nil {
		common.WriteError(c.Writer, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, refreshToken, err := auth.GenerateAuthTokens(u.Email)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	res := &pb.AuthResponse{
		AccessToken:	accessToken,
		RefreshToken:	refreshToken,
		CustomerID:	u.Email,
	}
	common.WriteJSON(c.Writer, http.StatusOK, res)
}

func (h *handler) handleRefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	customerID, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	accessToken, newRefreshToken, err := auth.GenerateAuthTokens(customerID)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	res := &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		CustomerID:   customerID,
	}
	common.WriteJSON(c.Writer, http.StatusOK, res)
}

func (h *handler) handleGetCart(c *gin.Context) {

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
		CustomerID:	customerID,
		Item:		&item,
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
		CustomerID:	customerID,
		ProductID:	productID,
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

func (h *handler) handleGetOrder(c *gin.Context) {
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

	common.WriteJSON(c.Writer, http.StatusOK, o)
}

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

func (h *handler) handleCreateOrder(c *gin.Context) {
	customerID := c.Param("customerID")

	if customerID != c.GetString("customerID") {
		common.WriteError(c.Writer, http.StatusForbidden, "not allowed to access this order")
		return
	}

	var items []*pb.ItemsWithQuantity
	if err := common.ReadJSON(c.Request, &items); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateItems(items); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	o, err := h.ordersGateway.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID:	customerID,
		Items:		items,
	})
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	res := &CreateOrderRequest{
		Order:		o,
		RedirectToURL:	fmt.Sprintf("%s/success.html?customerID=%s&orderID=%s", h.publicURL, o.CustomerID, o.ID),
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
