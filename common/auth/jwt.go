package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken	= errors.New("invalid token")

	ErrExpiredToken	= errors.New("expired token")
)

var jwtSecretKey = []byte("my-super-secret-key-1234")

type CustomClaims struct {
	CustomerID	string	`json:"customer_id"`
	jwt.RegisteredClaims
}

func GenerateToken(customerID string, duration time.Duration) (string, error) {

	claims := CustomClaims{
		CustomerID:	customerID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt:	jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:	jwt.NewNumericDate(time.Now()),
			NotBefore:	jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecretKey)
}

func ValidateToken(tokenString string) (string, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecretKey, nil
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

	return claims.CustomerID, nil
}
