package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("supersecretdevkey") // любой секрет, одинаковый для генерации и проверки

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(userID uint) (map[string]string, error) {
	accessToken := jwt.New(jwt.SigningMethodHS256)
	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["user_id"] = userID
	accessClaims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	accessClaims["iat"] = time.Now().Unix()
	accessClaims["iss"] = "ravell-api"

	accessString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["user_id"] = userID
	refreshClaims["exp"] = time.Now().Add(365 * 24 * time.Hour).Unix()
	refreshClaims["iat"] = time.Now().Unix()
	refreshClaims["iss"] = "ravell-api"

	refreshString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_token":  accessString,
		"refresh_token": refreshString,
	}, nil
}

func ValidateToken(tokenString string) (uint, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	return claims.UserID, nil
}
