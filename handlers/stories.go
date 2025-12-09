package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_stories_api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Функция подсчета слов
func countWords(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	return len(strings.Fields(text))
}

func GetStories(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	searchTerm := c.Query("search")
	query := db.Preload("User").Preload("User.Profile").Order("created_at DESC")

	if searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		query = query.Where(
			"title LIKE ? OR content LIKE ?", 
			searchPattern, 
			searchPattern,
		)
		log.Printf("Searching stories for term: %s", searchTerm)
	}

	var stories []models.Story
	if err := query.Find(&stories).Error; err != nil {
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
	if err := db.Preload("User").Preload("User.Profile").First(&story, id).Error; err != nil {
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
		ReplyTo  *uint  `json:"reply_to"` // ДОБАВЛЕНО
		Hashtags []uint `json:"hashtag_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка 100 слов
	wordCount := countWords(req.Content)
	if wordCount != 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Нужно ровно 100 слов. Сейчас: %d", wordCount),
		})
		return
	}

	story := models.Story{
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		WordCount: wordCount,
		ReplyTo:   req.ReplyTo,
	}

	// Транзакция для атомарности
	tx := db.Begin()
	
	if err := tx.Create(&story).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	// Если это ответ, обновляем родительскую историю
	if req.ReplyTo != nil {
		now := time.Now()
		if err := tx.Model(&models.Story{}).
			Where("id = ?", *req.ReplyTo).
			Updates(map[string]interface{}{
				"reply_count":   gorm.Expr("reply_count + 1"),
				"last_reply_at": now,
			}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update parent story"})
			return
		}
	}

	// Привязка хештегов
	for _, hashtagID := range req.Hashtags {
		var hashtag models.Hashtag
		if err := tx.First(&hashtag, hashtagID).Error; err != nil {
			continue // Пропускаем несуществующие хештеги
		}

		if err := tx.Create(&models.StoryHashtag{
			StoryID:   story.ID,
			HashtagID: hashtagID,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add hashtag"})
			return
		}
	}

	tx.Commit()

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
	log.Printf("Deleting story: %+v", story)

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

	var story models.Story
	if err := db.First(&story, storyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	var existingLike models.Like
	if err := db.Where("user_id = ? AND story_id = ?", userID, storyID).First(&existingLike).Error; err == nil {
		db.Delete(&existingLike)
	} else {
		db.Create(&models.Like{
			UserID:  userID,
			StoryID: uint(storyID),
		})
	}

	var likesCount int64
	db.Model(&models.Like{}).Where("story_id = ?", storyID).Count(&likesCount)

	c.JSON(http.StatusOK, gin.H{
		"liked":       err != nil,
		"message":     "Operation successful",
		"likes_count": likesCount,
	})
}

func NotInterestedStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)

	storyID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	var story models.Story
	if err := db.First(&story, storyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

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
	if err := db.Preload("User").Preload("User.Profile").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user stories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stories": stories,
		"count":   len(stories),
	})
}

func GetReplies(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)

    // 1. Получаем ID родительской истории из параметра URL
    parentID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
        return
    }

    // 2. Ищем все истории, где ReplyTo совпадает с ID родителя
    var replies []models.Story
    if err := db.Preload("User").Preload("User.Profile").
        Where("reply_to = ?", parentID). 
        Order("created_at ASC"). // Упорядочиваем по времени создания, чтобы видеть хронологию ответов
        Find(&replies).Error; err != nil {
        
        // В случае ошибки базы данных
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch replies"})
        return
    }
    
    // Опционально: Проверка, что родительская история существует
    // Хотя запрос Where("reply_to = ?", parentID) вернет пустой список,
    // если родителя нет, явная 404 может быть полезна. Но для списка ответов 
    // пустой список (200 OK) является стандартным поведением.

    // 3. Отправляем ответы
    c.JSON(http.StatusOK, gin.H{
        "replies": replies,
        "count":   len(replies),
    })
}

func GetSeeds(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var stories []models.Story
	if err := db.Preload("User").Preload("User.Profile").
		Where("reply_to IS NULL AND reply_count = 0").
		Order("created_at DESC").
		Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seeds"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"stories": stories})
}

func GetBranches(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var stories []models.Story
	if err := db.Preload("User").Preload("User.Profile").
		Where("reply_to IS NULL AND reply_count > 0").
		Order("reply_count DESC").
		Order("last_reply_at DESC").
		Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branches"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"stories": stories})
}