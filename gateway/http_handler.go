package main

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/commons/auth"
	"github.com/jaydeepjr28/commons/broker"
	"github.com/jaydeepjr28/omsv2-gateway/gateway"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	FrontendURL  string
}

type handler struct {
	ordersGateway gateway.OrdersGateway
	store         *store
	cartGateway   gateway.CartGateway
	publicURL     string
	producer      *broker.Producer
	rdb           *redis.Client
	oauthCfg      GoogleOAuthConfig
	adminEmail    string
}

func NewHandler(orders gateway.OrdersGateway, store *store, cart gateway.CartGateway, publicURL string, producer *broker.Producer, rdb *redis.Client, oauthCfg GoogleOAuthConfig, adminEmail string) *handler {
	return &handler{ordersGateway: orders, store: store, cartGateway: cart, publicURL: publicURL, producer: producer, rdb: rdb, oauthCfg: oauthCfg, adminEmail: adminEmail}
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

		customerID, role, err := auth.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("customerID", customerID)
		c.Set("role", role)
		c.Next()
	}
}

func adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}
		c.Next()
	}
}

func (h *handler) registerRoutes(router *gin.Engine) {
	router.POST("/api/users/register", h.handleRegister)
	router.POST("/api/users/register/send-otp", h.handleSendRegistrationOTP)
	router.POST("/api/users/register/verify-otp", h.handleVerifyRegistrationOTP)
	router.POST("/api/users/register/resend-otp", h.handleResendRegistrationOTP)
	router.POST("/api/users/login", h.handleLogin)
	router.POST("/api/users/refresh", h.handleRefreshToken)

	// Google OAuth 2.0
	router.GET("/api/auth/google/login", h.handleGoogleLogin)
	router.GET("/api/auth/google/callback", h.handleGoogleCallback)

	api := router.Group("/api")
	api.Use(authMiddleware())
	{
		api.POST("/customers/:customerID/orders", h.handleCreateOrder)
		api.GET("/customers/:customerID/orders", h.handleGetCustomerOrders)
		api.GET("/customers/:customerID/orders/:orderID", h.handleGetOrder)
		api.GET("/customers/:customerID/orders/:orderID/status", h.handleGetOrderStatus)
		api.POST("/customers/:customerID/orders/:orderID/verify-payment", h.handleVerifyOrderPayment)
		api.POST("/customers/:customerID/orders/:orderID/cancel", h.handleCancelOrder)
		api.GET("/inventory", h.handleGetInventory)

		api.GET("/cart", h.handleGetCart)
		api.POST("/cart/items", h.handleAddToCart)
		api.DELETE("/cart/items/:productID", h.handleRemoveFromCart)
		api.DELETE("/cart", h.handleClearCart)
	}

	admin := router.Group("/api/admin")
	admin.Use(authMiddleware(), adminAuthMiddleware())
	{
		admin.POST("/inventory", h.handleAdminCreateInventory)
		admin.PUT("/inventory/:id", h.handleAdminUpdateInventory)
		admin.GET("/orders", h.handleAdminGetOrders)
		admin.POST("/orders/:orderID/cancel", h.handleAdminCancelOrder)
		admin.GET("/analytics", h.handleAdminGetAnalytics)
	}
}

func (h *handler) handleRegister(c *gin.Context) {
	var req pb.RegisterRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	role := "customer"
	if req.Email == h.adminEmail || req.Email == "admin@stratum.com" || req.Email == "admin@oms.com" {
		role = "admin"
	}

	u := &UserModel{
		Email:		req.Email,
		Password:	req.Password,
		Name:		req.Name,
		Role:       role,
	}

	if err := h.store.CreateUser(c.Request.Context(), u); err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	accessToken, refreshToken, err := auth.GenerateAuthTokens(u.Email, u.Role)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	res := &pb.AuthResponse{
		AccessToken:	accessToken,
		RefreshToken:	refreshToken,
		CustomerID:	u.Email,
	}

	if h.producer != nil {
		body, _ := json.Marshal(res)
		_ = h.producer.Publish(c.Request.Context(), broker.UserRegisteredEvent, []byte(u.Email), body)
	}

	common.WriteJSON(c.Writer, http.StatusOK, res)
}

type RegistrationIntent struct {
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"`
	Role         string    `json:"role"`
	OTP          string    `json:"otp"`
	Attempts     int       `json:"attempts"`
	CreatedAt    time.Time `json:"createdAt"`
}

func generateSecureOTP() string {
	n, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(900000))
	if err != nil {
		return fmt.Sprintf("%06d", (time.Now().UnixNano()%900000)+100000)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

type sendOTPRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *handler) handleSendRegistrationOTP(c *gin.Context) {
	var req sendOTPRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, "Invalid request payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || !strings.Contains(req.Email, "@") {
		common.WriteError(c.Writer, http.StatusBadRequest, "A valid email address is required")
		return
	}
	if len(req.Name) < 2 {
		common.WriteError(c.Writer, http.StatusBadRequest, "Name must be at least 2 characters")
		return
	}
	if len(req.Password) < 6 {
		common.WriteError(c.Writer, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	ctx := c.Request.Context()

	// 1. Check if user already exists in PostgreSQL
	exists, err := h.store.EmailExists(ctx, req.Email)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, "Failed to verify account availability")
		return
	}
	if exists {
		common.WriteError(c.Writer, http.StatusConflict, "An account with this email address already exists. Please sign in.")
		return
	}

	// 2. Check 60-second cooldown in Redis
	cooldownKey := fmt.Sprintf("reg:cooldown:%s", req.Email)
	if h.rdb != nil {
		ttl, _ := h.rdb.TTL(ctx, cooldownKey).Result()
		if ttl > 0 {
			common.WriteError(c.Writer, http.StatusTooManyRequests, fmt.Sprintf("Please wait %d seconds before requesting another code.", int(ttl.Seconds())))
			return
		}
	}

	// 3. Generate secure OTP and bcrypt-hash password
	otp := generateSecureOTP()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, "Failed to encrypt password")
		return
	}

	role := "customer"
	if req.Email == h.adminEmail || req.Email == "admin@stratum.com" || req.Email == "admin@oms.com" {
		role = "admin"
	}

	intent := RegistrationIntent{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		OTP:          otp,
		Attempts:     0,
		CreatedAt:    time.Now(),
	}

	// 4. Save intent to Redis (5 min TTL) & set cooldown (60s TTL)
	if h.rdb != nil {
		intentData, _ := json.Marshal(intent)
		otpKey := fmt.Sprintf("reg:otp:%s", req.Email)
		if err := h.rdb.Set(ctx, otpKey, intentData, 5*time.Minute).Err(); err != nil {
			log.Printf("[Gateway] Warning: Failed to save OTP in Redis: %v", err)
			common.WriteError(c.Writer, http.StatusInternalServerError, "Failed to initiate verification session")
			return
		}
		_ = h.rdb.Set(ctx, cooldownKey, "1", 60*time.Second).Err()
	}

	// 5. Publish UserOTPRequestedEvent to Kafka
	if h.producer != nil {
		payload := broker.UserOTPRequestedPayload{
			Email:            req.Email,
			Name:             req.Name,
			OTP:              otp,
			ExpiresInSeconds: 300,
			RequestedAt:      time.Now(),
		}
		body, _ := json.Marshal(payload)
		_ = h.producer.Publish(ctx, broker.UserOTPRequestedEvent, []byte(req.Email), body)
	}

	common.WriteJSON(c.Writer, http.StatusOK, gin.H{
		"message":          "Verification code sent to your email",
		"email":            req.Email,
		"expiresInSeconds": 300,
		"cooldownSeconds":  60,
	})
}

type verifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (h *handler) handleVerifyRegistrationOTP(c *gin.Context) {
	var req verifyOTPRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, "Invalid request payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.OTP = strings.TrimSpace(req.OTP)

	if req.Email == "" || req.OTP == "" {
		common.WriteError(c.Writer, http.StatusBadRequest, "Email and 6-digit verification code are required")
		return
	}

	ctx := c.Request.Context()
	otpKey := fmt.Sprintf("reg:otp:%s", req.Email)

	if h.rdb == nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, "Verification cache unavailable")
		return
	}

	data, err := h.rdb.Get(ctx, otpKey).Bytes()
	if err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, "Verification code has expired or was not requested. Please request a new code.")
		return
	}

	var intent RegistrationIntent
	if err := json.Unmarshal(data, &intent); err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, "Failed to parse verification session")
		return
	}

	if intent.Attempts >= 3 {
		_ = h.rdb.Del(ctx, otpKey).Err()
		common.WriteError(c.Writer, http.StatusBadRequest, "Too many failed attempts. For security, please request a new verification code.")
		return
	}

	if intent.OTP != req.OTP {
		intent.Attempts++
		ttl, _ := h.rdb.TTL(ctx, otpKey).Result()
		if ttl > 0 {
			updatedData, _ := json.Marshal(intent)
			_ = h.rdb.Set(ctx, otpKey, updatedData, ttl).Err()
		}
		remaining := 3 - intent.Attempts
		common.WriteError(c.Writer, http.StatusBadRequest, fmt.Sprintf("Invalid verification code. %d attempt(s) remaining.", remaining))
		return
	}

	// Correct OTP! Create verified user in PostgreSQL
	u := &UserModel{
		Email:    intent.Email,
		Password: intent.PasswordHash,
		Name:     intent.Name,
		Role:     intent.Role,
	}

	if err := h.store.CreateVerifiedUser(ctx, u); err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, "Failed to create account: "+err.Error())
		return
	}

	// Clear session from Redis
	_ = h.rdb.Del(ctx, otpKey).Err()
	_ = h.rdb.Del(ctx, fmt.Sprintf("reg:cooldown:%s", req.Email)).Err()

	// Generate JWT auth tokens
	accessToken, refreshToken, err := auth.GenerateAuthTokens(u.Email, u.Role)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, "Account created, but failed to generate session token")
		return
	}

	res := &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CustomerID:   u.Email,
	}

	// Publish UserRegisteredEvent for welcome notification
	if h.producer != nil {
		body, _ := json.Marshal(res)
		_ = h.producer.Publish(ctx, broker.UserRegisteredEvent, []byte(u.Email), body)
	}

	common.WriteJSON(c.Writer, http.StatusCreated, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"customerID":   u.Email,
		"role":         u.Role,
		"name":         u.Name,
		"message":      "Email verified and account created successfully!",
	})
}

type resendOTPRequest struct {
	Email string `json:"email"`
}

func (h *handler) handleResendRegistrationOTP(c *gin.Context) {
	var req resendOTPRequest
	if err := common.ReadJSON(c.Request, &req); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, "Invalid request payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		common.WriteError(c.Writer, http.StatusBadRequest, "Email is required")
		return
	}

	ctx := c.Request.Context()
	cooldownKey := fmt.Sprintf("reg:cooldown:%s", req.Email)

	if h.rdb != nil {
		ttl, _ := h.rdb.TTL(ctx, cooldownKey).Result()
		if ttl > 0 {
			common.WriteError(c.Writer, http.StatusTooManyRequests, fmt.Sprintf("Please wait %d seconds before requesting another code.", int(ttl.Seconds())))
			return
		}
	}

	otpKey := fmt.Sprintf("reg:otp:%s", req.Email)
	data, err := h.rdb.Get(ctx, otpKey).Bytes()
	if err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, "Registration session not found or expired. Please fill out the registration form again.")
		return
	}

	var intent RegistrationIntent
	if err := json.Unmarshal(data, &intent); err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, "Failed to read registration session")
		return
	}

	// Generate fresh OTP
	newOTP := generateSecureOTP()
	intent.OTP = newOTP
	intent.Attempts = 0
	intent.CreatedAt = time.Now()

	updatedData, _ := json.Marshal(intent)
	_ = h.rdb.Set(ctx, otpKey, updatedData, 5*time.Minute).Err()
	_ = h.rdb.Set(ctx, cooldownKey, "1", 60*time.Second).Err()

	// Publish event to Kafka
	if h.producer != nil {
		payload := broker.UserOTPRequestedPayload{
			Email:            intent.Email,
			Name:             intent.Name,
			OTP:              newOTP,
			ExpiresInSeconds: 300,
			RequestedAt:      time.Now(),
		}
		body, _ := json.Marshal(payload)
		_ = h.producer.Publish(ctx, broker.UserOTPRequestedEvent, []byte(intent.Email), body)
	}

	common.WriteJSON(c.Writer, http.StatusOK, gin.H{
		"message":          "A fresh verification code has been dispatched to your email.",
		"email":            intent.Email,
		"expiresInSeconds": 300,
		"cooldownSeconds":  60,
	})
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

	accessToken, refreshToken, err := auth.GenerateAuthTokens(u.Email, u.Role)
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

	customerID, role, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	accessToken, newRefreshToken, err := auth.GenerateAuthTokens(customerID, role)
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

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func (h *handler) handleGoogleLogin(c *gin.Context) {
	redirectPath := c.DefaultQuery("redirect", "/")
	if !strings.HasPrefix(redirectPath, "/") {
		redirectPath = "/"
	}

	// 1. If Google credentials are not configured, render the interactive Developer Consent Sandbox
	if h.oauthCfg.ClientID == "" || h.oauthCfg.ClientSecret == "" {
		h.renderDevGoogleConsent(c, redirectPath)
		return
	}

	// 2. Generate random 16-byte hex state to mitigate CSRF
	stateBytes := make([]byte, 16)
	_, _ = cryptoRand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	// Save state -> redirectPath in Redis with 10 min TTL
	if h.rdb != nil {
		_ = h.rdb.Set(c.Request.Context(), "oauth:state:"+state, redirectPath, 10*time.Minute).Err()
	}

	// 3. Redirect to official Google OAuth 2.0 authorization endpoint
	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&access_type=online&prompt=select_account",
		url.QueryEscape(h.oauthCfg.ClientID),
		url.QueryEscape(h.oauthCfg.RedirectURL),
		url.QueryEscape("openid email profile"),
		url.QueryEscape(state),
	)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *handler) renderDevGoogleConsent(c *gin.Context, redirectPath string) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Google Sign-In (Development Mode) - STRATUM</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; }
    body { background-color: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; min-height: 100vh; padding: 20px; }
    .card { background: #1e293b; border: 1px solid #334155; border-radius: 16px; width: 100%%; max-width: 440px; padding: 32px; box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5); }
    .logo-container { display: flex; align-items: center; justify-content: center; gap: 12px; margin-bottom: 24px; }
    .badge { background: #3b82f6; color: white; padding: 4px 8px; border-radius: 6px; font-size: 11px; font-weight: 700; letter-spacing: 0.5px; text-transform: uppercase; }
    h1 { font-size: 20px; font-weight: 700; text-align: center; margin-bottom: 8px; }
    p.sub { font-size: 13px; color: #94a3b8; text-align: center; margin-bottom: 24px; line-height: 1.4; }
    .preset-btn { width: 100%%; display: flex; align-items: center; gap: 12px; background: #0f172a; border: 1px solid #334155; border-radius: 10px; padding: 12px 14px; margin-bottom: 12px; cursor: pointer; text-decoration: none; color: inherit; transition: all 0.2s; }
    .preset-btn:hover { border-color: #3b82f6; background: #1e293b; transform: translateY(-1px); }
    .avatar { width: 36px; height: 36px; border-radius: 50%%; background: #3b82f6; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 14px; color: white; }
    .avatar.admin { background: #8b5cf6; }
    .info { text-align: left; }
    .info .name { font-size: 14px; font-weight: 600; }
    .info .email { font-size: 12px; color: #94a3b8; }
    .divider { display: flex; align-items: center; text-align: center; margin: 20px 0; color: #64748b; font-size: 11px; text-transform: uppercase; letter-spacing: 1px; }
    .divider::before, .divider::after { content: ''; flex: 1; border-bottom: 1px solid #334155; }
    .divider span { padding: 0 10px; }
    form { display: flex; flex-direction: column; gap: 12px; }
    input { background: #0f172a; border: 1px solid #334155; border-radius: 8px; padding: 10px 14px; color: white; font-size: 13px; }
    input:focus { outline: none; border-color: #3b82f6; }
    button.submit-btn { background: #2563eb; color: white; border: none; border-radius: 8px; padding: 12px; font-size: 14px; font-weight: 600; cursor: pointer; transition: background 0.2s; }
    button.submit-btn:hover { background: #1d4ed8; }
    .footer-note { margin-top: 20px; font-size: 11px; color: #64748b; text-align: center; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo-container">
      <svg width="28" height="28" viewBox="0 0 24 24"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/></svg>
      <span class="badge">Dev Sandbox</span>
    </div>
    <h1>Sign in with Google</h1>
    <p class="sub">Choose a demo account or enter any Google identity to simulate OAuth verification instantly.</p>
    
    <a href="/api/auth/google/callback?code=mock_code&mock_name=Alex%%20Mercer&mock_email=alex.google@stratum.com&redirect=%s" class="preset-btn">
      <div class="avatar">A</div>
      <div class="info">
        <div class="name">Alex Mercer</div>
        <div class="email">alex.google@stratum.com (Customer)</div>
      </div>
    </a>

    <a href="/api/auth/google/callback?code=mock_code&mock_name=Operations%%20Lead&mock_email=%s&redirect=%s" class="preset-btn">
      <div class="avatar admin">★</div>
      <div class="info">
        <div class="name">Operations Lead</div>
        <div class="email">%s (Administrator)</div>
      </div>
    </a>

    <div class="divider"><span>Or custom account</span></div>

    <form action="/api/auth/google/callback" method="GET">
      <input type="hidden" name="code" value="mock_code" />
      <input type="hidden" name="redirect" value="%s" />
      <input type="text" name="mock_name" placeholder="Full Name (e.g. Sarah Connor)" required />
      <input type="email" name="mock_email" placeholder="Google Email (e.g. sarah@gmail.com)" required />
      <button type="submit" class="submit-btn">Authorize &amp; Continue</button>
    </form>

    <div class="footer-note">
      To use production Google OAuth, set <code>GOOGLE_CLIENT_ID</code> and <code>GOOGLE_CLIENT_SECRET</code> in <code>.env</code>.
    </div>
  </div>
</body>
</html>`, url.QueryEscape(redirectPath), url.QueryEscape(h.adminEmail), url.QueryEscape(redirectPath), h.adminEmail, redirectPath)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func (h *handler) handleGoogleCallback(c *gin.Context) {
	ctx := c.Request.Context()

	if errParam := c.Query("error"); errParam != "" {
		errDesc := c.DefaultQuery("error_description", "Authentication was cancelled")
		target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape(errDesc))
		c.Redirect(http.StatusTemporaryRedirect, target)
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" {
		target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Missing authorization code"))
		c.Redirect(http.StatusTemporaryRedirect, target)
		return
	}

	redirectPath := "/"
	if h.rdb != nil && state != "" {
		if val, err := h.rdb.Get(ctx, "oauth:state:"+state).Result(); err == nil && val != "" {
			redirectPath = val
			_ = h.rdb.Del(ctx, "oauth:state:"+state).Err()
		}
	}
	if overrideRedirect := c.Query("redirect"); overrideRedirect != "" && strings.HasPrefix(overrideRedirect, "/") {
		redirectPath = overrideRedirect
	}

	var googleUser GoogleUserInfo

	if strings.HasPrefix(code, "mock_") {
		email := strings.TrimSpace(strings.ToLower(c.DefaultQuery("mock_email", "google.user@example.com")))
		name := strings.TrimSpace(c.DefaultQuery("mock_name", "Google User"))
		picture := c.DefaultQuery("mock_picture", "")
		if email == "" {
			email = "google.user@example.com"
		}
		googleUser = GoogleUserInfo{
			ID:            fmt.Sprintf("mock_gid_%d", time.Now().Unix()),
			Email:         email,
			VerifiedEmail: true,
			Name:          name,
			Picture:       picture,
		}
	} else {
		tokenURL := "https://oauth2.googleapis.com/token"
		form := url.Values{
			"code":          {code},
			"client_id":     {h.oauthCfg.ClientID},
			"client_secret": {h.oauthCfg.ClientSecret},
			"redirect_uri":  {h.oauthCfg.RedirectURL},
			"grant_type":    {"authorization_code"},
		}

		client := &http.Client{Timeout: 10 * time.Second}
		tokenResp, err := client.PostForm(tokenURL, form)
		if err != nil {
			target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Failed to reach Google token endpoint"))
			c.Redirect(http.StatusTemporaryRedirect, target)
			return
		}
		defer tokenResp.Body.Close()

		if tokenResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(tokenResp.Body)
			log.Printf("[OAuth] Google token exchange failed (%d): %s", tokenResp.StatusCode, string(body))
			target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Google authorization code expired or invalid"))
			c.Redirect(http.StatusTemporaryRedirect, target)
			return
		}

		var tok GoogleTokenResponse
		if err := json.NewDecoder(tokenResp.Body).Decode(&tok); err != nil {
			target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Failed to decode token response"))
			c.Redirect(http.StatusTemporaryRedirect, target)
			return
		}

		userInfoReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		userInfoReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)

		userInfoResp, err := client.Do(userInfoReq)
		if err != nil || userInfoResp.StatusCode != http.StatusOK {
			target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Failed to fetch Google profile"))
			c.Redirect(http.StatusTemporaryRedirect, target)
			return
		}
		defer userInfoResp.Body.Close()

		if err := json.NewDecoder(userInfoResp.Body).Decode(&googleUser); err != nil {
			target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Failed to parse user profile"))
			c.Redirect(http.StatusTemporaryRedirect, target)
			return
		}
	}

	googleUser.Email = strings.TrimSpace(strings.ToLower(googleUser.Email))
	if googleUser.Email == "" {
		target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Google profile has no verified email"))
		c.Redirect(http.StatusTemporaryRedirect, target)
		return
	}
	if googleUser.Name == "" {
		parts := strings.Split(googleUser.Email, "@")
		googleUser.Name = parts[0]
	}

	role := "customer"
	if googleUser.Email == h.adminEmail || googleUser.Email == "admin@stratum.com" || googleUser.Email == "admin@oms.com" {
		role = "admin"
	}

	userModel := &UserModel{
		Email:      googleUser.Email,
		Name:       googleUser.Name,
		Role:       role,
		Provider:   "google",
		ProviderID: googleUser.ID,
		AvatarURL:  googleUser.Picture,
	}

	isNew, err := h.store.UpsertOAuthUser(ctx, userModel)
	if err != nil {
		log.Printf("[OAuth] UpsertOAuthUser error: %v", err)
		target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Failed to provision account"))
		c.Redirect(http.StatusTemporaryRedirect, target)
		return
	}

	accessToken, refreshToken, err := auth.GenerateAuthTokens(userModel.Email, userModel.Role)
	if err != nil {
		target := fmt.Sprintf("%s/login?error=%s", h.oauthCfg.FrontendURL, url.QueryEscape("Failed to create user session"))
		c.Redirect(http.StatusTemporaryRedirect, target)
		return
	}

	// Emit Kafka event if new user registered
	if isNew && h.producer != nil {
		res := &pb.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			CustomerID:   userModel.Email,
		}
		body, _ := json.Marshal(res)
		_ = h.producer.Publish(ctx, broker.UserRegisteredEvent, []byte(userModel.Email), body)
	}

	frontendCallback := fmt.Sprintf(
		"%s/auth/callback?access_token=%s&refresh_token=%s&role=%s&customer_id=%s&name=%s&avatar_url=%s&redirect=%s",
		h.oauthCfg.FrontendURL,
		url.QueryEscape(accessToken),
		url.QueryEscape(refreshToken),
		url.QueryEscape(userModel.Role),
		url.QueryEscape(userModel.Email),
		url.QueryEscape(userModel.Name),
		url.QueryEscape(userModel.AvatarURL),
		url.QueryEscape(redirectPath),
	)

	c.Redirect(http.StatusTemporaryRedirect, frontendCallback)
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
	if c.GetString("role") == "admin" {
		common.WriteError(c.Writer, http.StatusForbidden, "administrators are not permitted to add items to cart")
		return
	}

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

	if customerID != c.GetString("customerID") && c.GetString("role") != "admin" {
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

	if customerID != c.GetString("customerID") && c.GetString("role") != "admin" {
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

	if customerID != c.GetString("customerID") && c.GetString("role") != "admin" {
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

func (h *handler) handleVerifyOrderPayment(c *gin.Context) {
	customerID := c.Param("customerID")
	orderID := c.Param("orderID")

	if customerID != c.GetString("customerID") && c.GetString("role") != "admin" {
		common.WriteError(c.Writer, http.StatusForbidden, "not allowed to access this order")
		return
	}

	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	// Check if already paid or ready
	o, err := h.ordersGateway.GetOrder(ctx, orderID, customerID)
	if err == nil && (o.Status == "paid" || o.Status == "ready") {
		common.WriteJSON(c.Writer, http.StatusOK, gin.H{"status": o.Status, "verified": true})
		return
	}

	// Dispatch to payments HTTP service
	payload, _ := json.Marshal(gin.H{"orderID": orderID, "customerID": customerID})
	paymentsURL := common.EnvString("PAYMENTS_HTTP_ADDR", "http://localhost:8081") + "/verify-payment"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, paymentsURL, bytes.NewReader(payload))
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		common.WriteError(c.Writer, http.StatusBadGateway, "failed to reach payments service: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
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
	if c.GetString("role") == "admin" {
		common.WriteError(c.Writer, http.StatusForbidden, "administrators are not permitted to place customer orders")
		return
	}

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

func (h *handler) handleAdminCreateInventory(c *gin.Context) {
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	var req pb.Item
	if err := common.ReadJSON(c.Request, &req); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	if req.ID == "" {
		span.SetStatus(otelCodes.Error, "item ID is required")
		common.WriteError(c.Writer, http.StatusBadRequest, "item ID is required")
		return
	}
	if req.Name == "" {
		span.SetStatus(otelCodes.Error, "item Name is required")
		common.WriteError(c.Writer, http.StatusBadRequest, "item Name is required")
		return
	}
	if req.Quantity < 0 {
		span.SetStatus(otelCodes.Error, "item Quantity cannot be negative")
		common.WriteError(c.Writer, http.StatusBadRequest, "item Quantity cannot be negative")
		return
	}
	if req.Price < 100 {
		req.Price = 9900
	}

	created, err := h.ordersGateway.CreateItem(ctx, &req)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	common.WriteJSON(c.Writer, http.StatusCreated, created)
}

func (h *handler) handleAdminUpdateInventory(c *gin.Context) {
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	id := c.Param("id")
	if id == "" {
		span.SetStatus(otelCodes.Error, "item ID is required")
		common.WriteError(c.Writer, http.StatusBadRequest, "item ID is required")
		return
	}

	var req pb.Item
	if err := common.ReadJSON(c.Request, &req); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	req.ID = id
	if req.Quantity < 0 {
		span.SetStatus(otelCodes.Error, "item Quantity cannot be negative")
		common.WriteError(c.Writer, http.StatusBadRequest, "item Quantity cannot be negative")
		return
	}

	updated, err := h.ordersGateway.UpdateItem(ctx, &req)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	common.WriteJSON(c.Writer, http.StatusOK, updated)
}

func (h *handler) handleAdminGetOrders(c *gin.Context) {
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	res, err := h.ordersGateway.GetCustomerOrders(ctx, "")
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	orders := res.Orders
	if orders == nil {
		orders = []*pb.Order{}
	}
	common.WriteJSON(c.Writer, http.StatusOK, orders)
}

func (h *handler) handleAdminGetAnalytics(c *gin.Context) {
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	res, err := h.ordersGateway.GetCustomerOrders(ctx, "")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":         "admin get analytics",
			"orderVolume":     0,
			"statusBreakdown": map[string]int{},
		})
		return
	}

	statusMap := make(map[string]int)
	for _, o := range res.Orders {
		st := strings.ToLower(o.Status)
		if st == "" {
			st = "pending"
		}
		statusMap[st]++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "admin get analytics",
		"orderVolume":     len(res.Orders),
		"statusBreakdown": statusMap,
	})
}

func (h *handler) handleCancelOrder(c *gin.Context) {
	customerID := c.Param("customerID")
	orderID := c.Param("orderID")

	if customerID != c.GetString("customerID") && c.GetString("role") != "admin" {
		common.WriteError(c.Writer, http.StatusForbidden, "not allowed to cancel this order")
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

	if o.Status == "ready" {
		common.WriteError(c.Writer, http.StatusBadRequest, "order has already been prepared and completed")
		return
	}
	if o.Status == "cancelled" || o.Status == "refunded" {
		common.WriteError(c.Writer, http.StatusBadRequest, "order is already cancelled")
		return
	}

	marshalledOrder, err := json.Marshal(o)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	if h.producer != nil {
		if err := h.producer.Publish(ctx, broker.OrderCancelledEvent, []byte(o.ID), marshalledOrder); err != nil {
			common.WriteError(c.Writer, http.StatusInternalServerError, "failed to initiate cancellation saga: "+err.Error())
			return
		}
	}

	common.WriteJSON(c.Writer, http.StatusOK, gin.H{
		"message": "Cancellation Saga initiated successfully",
		"status":  "cancelled",
		"orderID": orderID,
	})
}

func (h *handler) handleAdminCancelOrder(c *gin.Context) {
	orderID := c.Param("orderID")

	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	o, err := h.ordersGateway.GetOrder(ctx, orderID, "")
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		writeGRPCError(c.Writer, err)
		return
	}

	marshalledOrder, err := json.Marshal(o)
	if err != nil {
		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	if h.producer != nil {
		if err := h.producer.Publish(ctx, broker.OrderCancelledEvent, []byte(o.ID), marshalledOrder); err != nil {
			common.WriteError(c.Writer, http.StatusInternalServerError, "failed to initiate cancellation saga: "+err.Error())
			return
		}
	}

	common.WriteJSON(c.Writer, http.StatusOK, gin.H{
		"message": "Admin cancellation Saga initiated successfully",
		"status":  "cancelled",
		"orderID": orderID,
	})
}


