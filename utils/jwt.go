package utils

import (
	"errors"
	"time"
	"ravell_backend/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(userID uint) (map[string]string, error) {
	cfg := config.LoadConfig()
	
	// Access token (24 hours)
	accessExpirationTime := time.Now().Add(24 * time.Hour)
	accessClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "stories-api",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Refresh token (7 days)
	refreshExpirationTime := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "stories-api",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
	}, nil
}

// ValidateToken проверяет JWT токен и возвращает userID
func ValidateToken(tokenString string) (uint, error) {
	cfg := config.LoadConfig()
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	return claims.UserID, nil
}

func RefreshToken(refreshToken string) (map[string]string, error) {
	userID, err := ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return GenerateJWTToken(userID)
}
