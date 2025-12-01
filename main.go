package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ravell_backend/database"
	"ravell_backend/handlers"
	"ravell_backend/middleware"
	"syscall"
	"time"

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
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		log.Println("Database connection closed")
	}()

	// Настройка Gin
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Создаем новый роутер
	r := gin.New()
	
	// ✅ Устанавливаем trusted proxies
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	
	// Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	
	// Логирование
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[GIN] %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Добавляем проверку соединения
	r.Use(func(c *gin.Context) {
		c.Next()
	})

	// Остальные middleware
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

	// 🏠 Health check с детальной информацией
	r.GET("/health", func(c *gin.Context) {
		// Проверяем подключение к БД
		sqlDB, err := db.DB()
		dbStatus := "connected"
		if err != nil {
			dbStatus = "error: " + err.Error()
		} else {
			if err := sqlDB.Ping(); err != nil {
				dbStatus = "error: " + err.Error()
			}
		}
		
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "Ravell API",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"database":  dbStatus,
			"environment": os.Getenv("ENV"),
			"host": c.Request.Host,
		})
	})

	// Простой root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Ravell Backend API v1.0.0",
			"health":  "/health",
			"docs":    "Coming soon",
		})
	})

	// ✅ СТАТИЧЕСКИЕ ФАЙЛЫ В КОНЦЕ
	r.Static("/media", "./media")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Изменил на 8080 для Render
	}

	// Создаем сервер с правильными настройками
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		log.Printf("🌐 Environment: %s", os.Getenv("ENV"))
		if os.Getenv("ENV") != "production" {
			log.Printf("📝 Debug mode enabled")
		}
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}
	
	log.Println("✅ Server exited properly")
}