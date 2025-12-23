package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"go_stories_api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SeedAchievements добавляет начальные ачивки
func SeedAchievements(db *gorm.DB) {
	achievements := []models.Achievement{
		{
			Key:         "early_access",
			Title:       "Первооткрыватель",
			Description: "Войти под ранний доступ программы",
			IconURL:     "https://cdn.ravell.app/achievements/early_access.png",
		},
	}

	for _, ach := range achievements {
		var existing models.Achievement
		// ищем по ключу, но не делаем db.Find
		if err := db.Where("key = ?", ach.Key).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&ach)
			} else {
				log.Printf("Ошибка при проверке ачивки %s: %v", ach.Key, err)
			}
		}
	}
}

// InitDB инициализирует подключение к базе данных
func InitDB() *gorm.DB {
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
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
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
	// делаем только AutoMigrate без каких-либо Find
	if err := db.AutoMigrate(
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
	); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	// Seed ачивки безопасно
	SeedAchievements(db)

	log.Println("✅ Database migration completed")
}
