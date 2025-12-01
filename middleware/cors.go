// middleware/cors.go - ПОЛНОСТЬЮ ПЕРЕПИШИ!
package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ⚠️ ВСЕГДА СТАВЬ ЗВЁЗДОЧКУ НА RENDER!
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "*")
		
		// ⚠️ ОБЯЗАТЕЛЬНО ДОБАВЬ ЭТО!
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		
		c.Next()
	}
}