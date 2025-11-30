package database

import (
	"log"
	"ravell_backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	dsn := "host=dpg-d4lhvlk9c44c73fhpnv0-a.oregon-postgres.render.com user=mydjangodb_p5sh_user password=l4JYUoXYOzMAjBxpN3yoe5OCV5qAbTMi dbname=mydjangodb_p5sh port=5432 sslmode=require"
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully!")
	return db
}

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
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("✅ Database migrated successfully")
}

func GetDB() *gorm.DB {
	return DB
}
