package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken	= errors.New("invalid token")

	ErrExpiredToken	= errors.New("expired token")
)

func getJWTSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	return []byte(secret)
}

type CustomClaims struct {
	CustomerID	string	`json:"customer_id"`
	TokenType	string	`json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateAuthTokens(customerID string) (string, string, error) {
	accessClaims := CustomClaims{
		CustomerID: customerID,
		TokenType:  "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessSignature, err := accessToken.SignedString(getJWTSecretKey())
	if err != nil {
		return "", "", err
	}

	refreshClaims := CustomClaims{
		CustomerID: customerID,
		TokenType:  "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSignature, err := refreshToken.SignedString(getJWTSecretKey())
	if err != nil {
		return "", "", err
	}

	return accessSignature, refreshSignature, nil
}

func validateTokenBase(tokenString string, expectedType string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecretKey(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return "", ErrInvalidToken
	}

	return claims.CustomerID, nil
}

func ValidateAccessToken(tokenString string) (string, error) {
	return validateTokenBase(tokenString, "access")
}

func ValidateRefreshToken(tokenString string) (string, error) {
	return validateTokenBase(tokenString, "refresh")
}
