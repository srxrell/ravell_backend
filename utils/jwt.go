package utils

import (
	"errors"
	"time"
	"os"
	"go_stories_api/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(userID uint) (map[string]string, error) {
	// Access token - 24 часа вместо 15 минут
	accessToken := jwt.New(jwt.SigningMethodHS256)
	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["user_id"] = userID
	accessClaims["exp"] = time.Now().Add(24 * time.Hour).Unix() // 🟢 24 часа
	accessClaims["iat"] = time.Now().Unix()
	accessClaims["iss"] = "ravell-api"

	accessString, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	// Refresh token - 30 дней
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["user_id"] = userID
	refreshClaims["exp"] = time.Now().Add(365 * 24 * time.Hour).Unix() // 🟢 30 дней
	refreshClaims["iat"] = time.Now().Unix()
	refreshClaims["iss"] = "ravell-api"

	refreshString, err := refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_token":  accessString,
		"refresh_token": refreshString,
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
