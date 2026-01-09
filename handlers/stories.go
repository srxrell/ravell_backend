package handlers

import (
	"fmt"
	"go_stories_api/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func RegisterView(db *gorm.DB, postId int, userId uint) error {
    query := `
        INSERT INTO post_views (post_id, user_id) 
        VALUES (?, ?) 
        ON CONFLICT (post_id, user_id) DO NOTHING`
    
    result := db.Exec(query, postId, userId)
    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected > 0 {
        err := db.Model(&models.Story{}).Where("id = ?", postId).
            Update("views", gorm.Expr("views + 1")).Error
        return err
    }

    return nil
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

	currentUserID := c.MustGet("user_id").(uint)

	go func(database *gorm.DB, pID int, uID uint) {
		if err := RegisterView(database, pID, uID); err != nil {
			log.Printf("Background view error: %v", err)
		}
	}(db, id, currentUserID)
	c.JSON(http.StatusOK, story)
}


func CreateStory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)

	var req struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		ReplyTo  *uint  `json:"reply_to"`
		Hashtags []uint `json:"hashtag_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка 100 слов
	wordCount := countWords(req.Content)
	// разрешаем истории с > 20 <= 100 слов
	if wordCount < 20 || wordCount > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Нужно от 20 до 100 слов. Сейчас: %d", wordCount),
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

	tx := db.Begin()
	if err := tx.Create(&story).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	// Привязка хештегов
	for _, hashtagID := range req.Hashtags {
		var hashtag models.Hashtag
		if err := tx.First(&hashtag, hashtagID).Error; err != nil {
			continue
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

	// Загружаем автора, чтобы использовать username в пушах
	db.Preload("User").Preload("User.Profile").First(&story, story.ID)

	// --- Пуш подписчикам автора ---
	var subs []models.Subscription
	db.Where("following_id = ?", userID).Find(&subs)

	var playerIDs []string
	for _, sub := range subs {
		var devices []models.UserDevice
		db.Where("user_id = ?", sub.FollowerID).Find(&devices)
		for _, d := range devices {
			if d.PlayerID != "" {
				playerIDs = append(playerIDs, d.PlayerID)
			}
		}
	}

	if len(playerIDs) > 0 {
		
	}

	// --- Пуш автору родительской истории, если это ответ ---
	if req.ReplyTo != nil {
		var parent models.Story
		if err := db.Preload("User").First(&parent, *req.ReplyTo).Error; err == nil {
			var devices []models.UserDevice
			db.Where("user_id = ?", parent.UserID).Find(&devices)
			var replyPlayerIDs []string
			for _, d := range devices {
				if d.PlayerID != "" {
					replyPlayerIDs = append(replyPlayerIDs, d.PlayerID)
				}
			}
			if len(replyPlayerIDs) > 0 {
				
			}

			// Обновляем родительскую историю
			now := time.Now()
			db.Model(&models.Story{}).Where("id = ?", parent.ID).Updates(map[string]interface{}{
				"reply_count":   gorm.Expr("reply_count + 1"),
				"last_reply_at": now,
			})
		}
	}

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

	// if story.UserID != userID {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Not your story"})
	// 	return
	// }

	// Транзакция для безопасного удаления
	tx := db.Begin()

	// 1️⃣ Удаляем связи с хештегами
	if err := tx.Where("story_id = ?", story.ID).Delete(&models.StoryHashtag{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove story hashtags links"})
		return
	}

	// 2️⃣ Удаляем саму историю
	if err := tx.Delete(&story).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete story"})
		return
	}

	tx.Commit()

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