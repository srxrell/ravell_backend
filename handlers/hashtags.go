package handlers

import (
	"net/http"
	"strconv"
	"go_stories_api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
