// database/database.go
package database

import (
	"go_stories_api/models"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB инициализирует подключение к базе данных
func InitDB() *gorm.DB {
	// Получаем строку подключения из переменных окружения
	dsn := "host=dpg-d4lhvlk9c44c73fhpnv0-a.oregon-postgres.render.com user=mydjangodb_p5sh_user password=l4JYUoXYOzMAjBxpN3yoe5OCV5qAbTMi dbname=mydjangodb_p5sh port=5432 sslmode=require"
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Настраиваем логгер для GORM
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

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Получаем объект sql.DB для настройки пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Настраиваем пул соединений
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Проверяем подключение
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

	)
	
	if err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}
	
	log.Println("✅ Database migration completed")
	
	// ПРОВЕРЬ, ЧТО ТАБЛИЦЫ СОЗДАНЫ
	var tableCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	log.Printf("📊 Tables created: %d", tableCount)
}