package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when a parsed JWT token fails validation checks.
	ErrInvalidToken = errors.New("invalid token")
	
	// ErrExpiredToken is returned when a parsed JWT token has passed its expiration time.
	ErrExpiredToken = errors.New("expired token")
)

// jwtSecretKey is the symmetric key used for HMAC signing of the JWT tokens.
// In a production environment, this MUST be loaded from a secure environment variable or secrets manager.
var jwtSecretKey = []byte("my-super-secret-key-1234")

// CustomClaims represents the payload embedded in the JWT token.
// It includes standard RegisteredClaims (like expiry, issue time) and custom application data.
type CustomClaims struct {
	CustomerID string `json:"customer_id"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT string for the given customer ID.
// The token is signed using HS256 and will be valid for the specified duration.
func GenerateToken(customerID string, duration time.Duration) (string, error) {
	// Construct the payload with the user's ID and standard time-based claims
	claims := CustomClaims{
		CustomerID: customerID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)), // Token expiration time
			IssuedAt:  jwt.NewNumericDate(time.Now()),               // When the token was issued
			NotBefore: jwt.NewNumericDate(time.Now()),               // Token is immediately valid
		},
	}

	// Create a new token object specifying the signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Cryptographically sign the token using our secret key and return the string representation
	return token.SignedString(jwtSecretKey)
}

// ValidateToken parses a JWT string, validates its signature, checks its expiration,
// and extracts the customerID if it is completely valid.
func ValidateToken(tokenString string) (string, error) {
	// Parse the token and pass a keyfunc that provides the secret key for signature verification
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the token's signing method is what we expect (HMAC) to prevent downgrade attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecretKey, nil
	})

	// Handle parsing and validation errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", err
	}

	// Assert that the parsed claims match our CustomClaims structure
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	// Return the securely extracted customer ID
	return claims.CustomerID, nil
}

