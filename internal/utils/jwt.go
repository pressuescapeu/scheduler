package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, email string) (string, error) {
	expireHours := os.Getenv("JWT_EXPIRY_HOURS")
	signingKey := os.Getenv("JWT_SECRET")

	if signingKey == "" {
		return "", errors.New("JWT_SECRET not set")
	}

	expirationHours, err := strconv.Atoi(expireHours)

	if err != nil {
		return "", errors.New("invalid JWT_EXPIRY_HOURS")
	}

	claims := Claims{
		userID,
		email,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(signingKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	signingKey := os.Getenv("JWT_SECRET")

	if signingKey == "" {
		return nil, errors.New("JWT_SECRET not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}
