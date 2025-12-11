// middleware/cors.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Разрешаем запросы от любого источника (звездочка *)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		
		// Разрешаем все заголовки, включая Content-Type, Authorization, и т.д.
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		
		// Разрешаем все основные HTTP методы
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		
		// Максимальное время кэширования preflight-запроса (в секундах)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Обязательная обработка OPTIONS запросов (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // Используем 204 No Content для OPTIONS
			return
		}
		
		c.Next()
	}
}