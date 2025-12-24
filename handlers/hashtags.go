package handlers

import (
	"net/http"
	"strconv"
	"go_stories_api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DeleteHashtag(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Получаем ID из параметров пути /hashtags/:id
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}

	// Сначала проверяем, существует ли такой хештег
	var hashtag models.Hashtag
	if err := db.First(&hashtag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Если ID=3 не найден, выдаст этот статус
			c.JSON(http.StatusNotFound, gin.H{"error": "Hashtag with this ID not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Используем транзакцию для безопасности данных
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Удаляем связи в связующей таблице (story_hashtags)
		// Это предотвратит ошибку Foreign Key Constraint
		if err := tx.Exec("DELETE FROM story_hashtags WHERE hashtag_id = ?", id).Error; err != nil {
			return err
		}

		// 2. Удаляем сам хештег
		if err := tx.Delete(&hashtag).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hashtag and its relations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hashtag deleted successfully",
		"id":      id,
		"name":    hashtag.Name,
	})
}


func GetHashtags(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var hashtags []models.Hashtag
	result := db.Order("name ASC").Find(&hashtags)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hashtags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hashtags": hashtags,
		"count":    len(hashtags),
	})
}

func CreateHashtag(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование хештега
	var existingHashtag models.Hashtag
	if err := db.Where("name = ?", req.Name).First(&existingHashtag).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Hashtag already exists"})
		return
	}

	hashtag := models.Hashtag{
		Name: req.Name,
	}

	if err := db.Create(&hashtag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hashtag"})
		return
	}

	c.JSON(http.StatusCreated, hashtag)
}

func GetHashtagStories(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	hashtagID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hashtag ID"})
		return
	}

	var hashtag models.Hashtag
	if err := db.First(&hashtag, hashtagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hashtag not found"})
		return
	}

	var stories []models.Story
	result := db.Joins("JOIN story_hashtags ON story_hashtags.story_id = stories.id").
		Where("story_hashtags.hashtag_id = ?", hashtagID).
		Preload("User").
		Preload("User.Profile").
		Order("stories.created_at DESC").
		Find(&stories)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hashtag": hashtag,
		"stories": stories,
		"count":   len(stories),
	})
}
