package main

import (
	"log"
	"os"
	"ravell_backend/database"
	"ravell_backend/handlers"
	"ravell_backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Инициализация базы данных
	db := database.InitDB()
	
	// Автомиграция
	database.MigrateDB(db)

	// Настройка Gin
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.DatabaseMiddleware(db))

	// 🔐 Аутентификация
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/verify-otp", handlers.VerifyOTP)
		auth.POST("/refresh-token", handlers.RefreshToken)
		auth.POST("/resend-otp", handlers.ResendOTP)
	}

	// 📖 Истории
	stories := r.Group("/api/stories")
	{
		stories.GET("/", handlers.GetStories)
		stories.GET("/:id", handlers.GetStory)
		stories.POST("/", handlers.CreateStory) // Убрал JWTAuth
		stories.PUT("/:id", handlers.UpdateStory) // Убрал JWTAuth
		stories.DELETE("/:id", handlers.DeleteStory) // Убрал JWTAuth
		stories.POST("/:id/like", handlers.LikeStory) // Убрал JWTAuth
		stories.POST("/:id/not-interested", handlers.NotInterestedStory) // Убрал JWTAuth
		stories.GET("/:id/comments", handlers.GetComments)
	}

	// 💬 Комментарии
	comments := r.Group("/api/comments") // Убрал JWTAuth
	{
		comments.POST("/", handlers.CreateComment)
		comments.PUT("/:id", handlers.UpdateComment)
		comments.DELETE("/:id", handlers.DeleteComment)
	}

	// 👥 Пользователи
	users := r.Group("/api/users")
	{
		users.GET("/:id/profile", handlers.GetUserProfile)
		users.GET("/:id/stories", handlers.GetUserStories)
		users.GET("/:id/followers", handlers.GetFollowers)
		users.GET("/:id/following", handlers.GetFollowing)
		users.POST("/:id/follow", handlers.FollowUser) // Убрал JWTAuth
		users.POST("/:id/unfollow", handlers.UnfollowUser) // Убрал JWTAuth
	}

	// 👤 Профиль
	profile := r.Group("/api/profile") // Убрал JWTAuth
	{
		profile.GET("/", handlers.GetMyProfile)
		profile.PUT("/", handlers.UpdateProfile)
		// Убрал UpdateAvatar
	}

	// 🏷️ Хештеги
	hashtags := r.Group("/api/hashtags")
	{
		hashtags.GET("/", handlers.GetHashtags)
		hashtags.GET("/:id/stories", handlers.GetHashtagStories)
		hashtags.POST("/", handlers.CreateHashtag) // Убрал JWTAuth
	}

	// 🏠 Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "Stories API",
			"version":   "1.0.0",
			"timestamp": gin.H{"server": "online", "database": "connected"},
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("🚀 Server started on port %s", port)
	log.Printf("📚 API Documentation: http://localhost:%s/health", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
