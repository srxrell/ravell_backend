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

	// ✅ ДОБАВЛЯЕМ СТАТИЧЕСКУЮ РАЗДАЧУ ФАЙЛОВ
	r.Static("/media", "./media")

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.DatabaseMiddleware(db))

	// 🔐 Аутентификация (прямые маршруты)
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.POST("/refresh-token", handlers.RefreshToken)

	// 👤 Профиль (защищенные маршруты)
	profile := r.Group("/")
	profile.Use(middleware.JWTAuth())
	{
		profile.GET("/profile", handlers.GetMyProfile)
		profile.PUT("/profile", handlers.UpdateProfile)
		profile.PUT("/profile/with-image", handlers.UpdateProfileWithImage)
		profile.DELETE("/account", handlers.DeleteAccount)
	}

	// 📖 Истории
	stories := r.Group("/stories")
	{
		stories.GET("/", handlers.GetStories)
		stories.GET("/:id", handlers.GetStory)
		stories.GET("/:id/comments", handlers.GetComments)
		
		// Защищенные маршруты для историй
		protectedStories := stories.Group("")
		protectedStories.Use(middleware.JWTAuth())
		{
			protectedStories.POST("/", handlers.CreateStory)
			protectedStories.PUT("/:id", handlers.UpdateStory)
			protectedStories.DELETE("/:id", handlers.DeleteStory)
			protectedStories.POST("/:id/like", handlers.LikeStory)
			protectedStories.POST("/:id/not-interested", handlers.NotInterestedStory)
		}
	}

	// 💬 Комментарии
	comments := r.Group("/comments")
	comments.Use(middleware.JWTAuth())
	{
		comments.POST("/", handlers.CreateComment)
		comments.PUT("/:id", handlers.UpdateComment)
		comments.DELETE("/:id", handlers.DeleteComment)
	}

	// 👥 Пользователи
	users := r.Group("/users")
	{
		users.GET("/:id/profile", handlers.GetUserProfile)
		users.GET("/:id/stories", handlers.GetUserStories)
		users.GET("/:id/followers", handlers.GetFollowers)
		users.GET("/:id/following", handlers.GetFollowing)
		
		// Защищенные маршруты для подписок
		protectedUsers := users.Group("")
		protectedUsers.Use(middleware.JWTAuth())
		{
			protectedUsers.POST("/:id/follow", handlers.FollowUser)
			protectedUsers.POST("/:id/unfollow", handlers.UnfollowUser)
		}
	}

	// 🏷️ Хештеги
	hashtags := r.Group("/hashtags")
	{
		hashtags.GET("/", handlers.GetHashtags)
		hashtags.GET("/:id/stories", handlers.GetHashtagStories)
		
		// Защищенные маршруты для хештегов
		protectedHashtags := hashtags.Group("")
		protectedHashtags.Use(middleware.JWTAuth())
		{
			protectedHashtags.POST("/", handlers.CreateHashtag)
		}
	}

	// 🏠 Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "Ravell API",
			"version":   "1.0.0",
			"timestamp": gin.H{"server": "online", "database": "connected"},
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Ravell Backend API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"auth":      "/register, /login, /refresh-token",
				"profile":   "/profile, /profile/with-image (protected)",
				"stories":   "/stories, /stories/:id",
				"comments":  "/comments (protected)", 
				"users":     "/users/:id/profile, /users/:id/stories",
				"hashtags":  "/hashtags, /hashtags/:id/stories",
				"health":    "/health",
			},
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("🚀 Server started on port %s", port)
	log.Printf("🌐 Base URL: https://ravell-backend-1.onrender.com")
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}