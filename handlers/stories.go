package handlers

import (
	"net/http"
	"strconv"
	"go_stories_api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetStories(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var stories []models.Story
	result := db.Preload("User").Preload("User.Profile").
		Order("created_at DESC").
		Find(&stories)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stories": stories,
		"count":   len(stories),
	})
}

func GetStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	var story models.Story
	result := db.Preload("User").Preload("User.Profile").First(&story, id)
	
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	c.JSON(http.StatusOK, story)
}

func CreateStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)
	
	var req struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Hashtags []uint `json:"hashtag_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	story := models.Story{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := db.Create(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	// Привязка хештегов
	for _, hashtagID := range req.Hashtags {
		var hashtag models.Hashtag
		if err := db.First(&hashtag, hashtagID).Error; err != nil {
			continue // Пропускаем несуществующие хештеги
		}
		
		db.Create(&models.StoryHashtag{
			StoryID:   story.ID,
			HashtagID: hashtagID,
		})
	}

	// Загружаем связанные данные для ответа
	db.Preload("User").Preload("User.Profile").First(&story, story.ID)

	c.JSON(http.StatusCreated, story)
}

func UpdateStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)
	
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	var story models.Story
	if err := db.First(&story, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	// Проверка владельца
	if story.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your story"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только переданные поля
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}

	if err := db.Model(&story).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update story"})
		return
	}

	c.JSON(http.StatusOK, story)
}

func DeleteStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)
	
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	var story models.Story
	if err := db.First(&story, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	if story.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your story"})
		return
	}

	if err := db.Delete(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete story"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Story deleted successfully"})
}

func LikeStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)
	
	storyID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	// Проверяем существование истории
	var story models.Story
	if err := db.First(&story, storyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	var existingLike models.Like
	if err := db.Where("user_id = ? AND story_id = ?", userID, storyID).First(&existingLike).Error; err == nil {
		// Удаляем лайк если существует
		db.Delete(&existingLike)
		c.JSON(http.StatusOK, gin.H{"liked": false, "message": "Like removed"})
	} else {
		// Создаем новый лайк
		db.Create(&models.Like{
			UserID:  userID,
			StoryID: uint(storyID),
		})
		c.JSON(http.StatusOK, gin.H{"liked": true, "message": "Story liked"})
	}
}

func NotInterestedStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)
	
	storyID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	// Проверяем существование истории
	var story models.Story
	if err := db.First(&story, storyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	// Добавляем в "Не интересно"
	notInterested := models.NotInterested{
		UserID:  userID,
		StoryID: uint(storyID),
	}

	if err := db.Create(&notInterested).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark story as not interested"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Story marked as not interested"})
}

func GetUserStories(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var stories []models.Story
	result := db.Preload("User").Preload("User.Profile").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&stories)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user stories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stories": stories,
		"count":   len(stories),
	})
}
