package handlers

import (
	"net/http"
	"strconv"
	"go_stories_api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DeleteHashtagByName(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Получаем имя из URL
	hashtagName := c.Param("name")
	if hashtagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hashtag name required"})
		return
	}

	// Ищем хештег
	var hashtag models.Hashtag
	if err := db.Where("name = ?", hashtagName).First(&hashtag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hashtag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hashtag"})
		}
		return
	}

	// Удаляем связи с историями
	if err := db.Exec("DELETE FROM story_hashtags WHERE hashtag_id = ?", hashtag.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove hashtag relations"})
		return
	}

	// Удаляем сам хештег
	if err := db.Delete(&hashtag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hashtag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hashtag deleted successfully"})
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
