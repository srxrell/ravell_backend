package main

import (
	"context"
	"fmt"
	"go_stories_api/database"
	"go_stories_api/handlers"
	"go_stories_api/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// ================= ENV =================
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// ================= DB =================
	db := database.InitDB()
	database.MigrateDB(db)
	database.SeedAchievements(db)
	database.SeedUserAchievements(db)


	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		log.Println("Database connection closed")
	}()

	// ================= GIN =================
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.DatabaseMiddleware(db))

	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf(
			"[GIN] %s [%s] \"%s %s\" %d %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	// ================= AUTH =================
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.POST("/refresh-token", handlers.RefreshToken)

	// =============== ACHIEVEMENTS ============


	// ================= PROFILE =================
	profile := r.Group("/")
	profile.Use(middleware.JWTAuth())
	{
		profile.GET("/profile", handlers.GetMyProfile)
		profile.PUT("/profile", handlers.UpdateProfile)
		profile.PUT("/profile/with-image", handlers.UpdateProfileWithImage)
		profile.DELETE("/account", handlers.DeleteAccount)
	}

	// ================= STORIES =================
	stories := r.Group("/stories")
	{
		stories.GET("/", handlers.GetStories)
		stories.GET("/seeds", handlers.GetSeeds)
		stories.GET("/branches", handlers.GetBranches)
		stories.GET("/:id", handlers.GetStory)
		stories.GET("/:id/comments", handlers.GetComments)
		stories.GET("/:id/replies", handlers.GetReplies)

		protected := stories.Group("/")
		protected.Use(middleware.JWTAuth())
		{
			protected.POST("/", handlers.CreateStory)
			protected.PUT("/:id", handlers.UpdateStory)
			protected.DELETE("/:id", handlers.DeleteStory)
			protected.POST("/:id/like", handlers.LikeStory)
			protected.POST("/:id/not-interested", handlers.NotInterestedStory)
		}
	}

	// ================= COMMENTS =================
	comments := r.Group("/comments")
	comments.Use(middleware.JWTAuth())
	{
		comments.GET("/all", handlers.GetAllComments)
		comments.POST("/", handlers.CreateComment)
		comments.PUT("/:id", handlers.UpdateComment)
		comments.DELETE("/:id", handlers.DeleteComment)
	}

	// ================= USERS =================
	users := r.Group("/users")
	{
		users.GET("/:id/profile", handlers.GetUserProfile)
		users.GET("/:id/stories", handlers.GetUserStories)
		users.GET("/:id/followers", handlers.GetFollowers)
		users.GET("/:id/following", handlers.GetFollowing)
		users.GET("/:id/streak", handlers.GetUserStreak)
		users.GET("/:id/achievements", middleware.JWTAuth(), handlers.GetUserAchievementsByID)


		protected := users.Group("/")
		protected.Use(middleware.JWTAuth())
		{
			protected.POST("/:id/follow", handlers.FollowUser)
			protected.POST("/:id/unfollow", handlers.UnfollowUser)
			protected.POST("/save-player", handlers.SavePlayerID)
		}
		users.GET("/influencers/early", handlers.GetActiveInfluencers)

	}

	streak := r.Group("/streak")
	streak.Use(middleware.JWTAuth())
	{
		streak.POST("/update", handlers.UpdateStreak)
		streak.GET("", handlers.GetStreak)
	}

	// ================= HASHTAGS =================
	hashtags := r.Group("/hashtags")
	{
		hashtags.GET("/", handlers.GetHashtags)
		hashtags.GET("/:id/stories", handlers.GetHashtagStories)

		protected := hashtags.Group("/")
		protected.Use(middleware.JWTAuth())
		{
			protected.POST("/", handlers.CreateHashtag)
		}
	}

	// ================= WS =================
	ws := r.Group("/ws")
	ws.Use(middleware.WSJWTAuth())
	{
		ws.GET("/", handlers.WSHandler)
	}

	// ================= MEDIA =================
	r.Static("/media", "./media")

	// ================= HEALTH =================
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "Ravell API",
			"time":    time.Now().Unix(),
		})
	})

	// ================= SERVER =================
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Println("🚀 Server started on port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// ================= SHUTDOWN =================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("✅ Server stopped")
}
