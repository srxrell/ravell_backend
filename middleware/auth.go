package middleware

import (
	"go_stories_api/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth проверяет Authorization header и сохраняет user_id в контексте
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]
		userID, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		// Вечный токен — сохраняем user_id как uint
		c.Set("user_id", userID)
		c.Next()
	}
}

// WSJWTAuth для WebSocket, тоже вечный токен
func WSJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "No token"})
			return
		}

		userID, err := utils.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalJWTAuth пытается получить user_id из токена, но не прерывает запрос если его нет
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		userID, err := utils.ValidateToken(token)
		if err != nil {
			// Если токен невалиден, просто продолжаем как гость
			c.Next()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
