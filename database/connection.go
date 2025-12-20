package database

import (
	"log"
	"os"
	"time"

	"go_stories_api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB инициализирует подключение к базе данных
func InitDB() *gorm.DB {
	dsn := "host=dpg-d4lhvlk9c44c73fhpnv0-a.oregon-postgres.render.com user=mydjangodb_p5sh_user password=l4JYUoXYOzMAjBxpN3yoe5OCV5qAbTMi dbname=mydjangodb_p5sh port=5432 sslmode=require"

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Database connection established")
	return db
}

// MigrateDB выполняет миграции
func MigrateDB(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.Story{},
		&models.Comment{},
		&models.Like{},
		&models.Subscription{},
		&models.Hashtag{},
		&models.StoryHashtag{},
		&models.NotInterested{},
		&models.UserDevice{},
		&models.Feature{},
		&models.Achievement{},
		&models.UserAchievement{},
	)

	if err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	log.Println("✅ Database migration completed")
}
