package handlers

import (
	"encoding/json"
	"go_stories_api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Возвращаем ачивки пользователя, создавая их на лету если их нет
func GetUserAchievementsByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idParam := c.Param("id")

	userID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var userAchievements []models.UserAchievement
	err = db.Preload("Achievement").Where("user_id = ?", userID).Find(&userAchievements).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	// Если нет записей — создаем на лету
	if len(userAchievements) == 0 {
		var allAchievements []models.Achievement
		db.Find(&allAchievements)

		for _, a := range allAchievements {
			userAchievements = append(userAchievements, models.UserAchievement{
				UserID:        uint(userID),
				AchievementID: a.ID,
				Progress:      0,
				Unlocked:      false,
				Achievement:   a,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"achievements": userAchievements})
}

// Обновление прогресса всех ачивок для всех пользователей
func UpdateAllAchievements(db *gorm.DB) {
	var achievements []models.Achievement
	db.Find(&achievements)

	var users []models.User
	db.Find(&users)

	for _, ach := range achievements {
		for _, user := range users {
			progress := calculateProgress(db, user.ID, ach)

			// Ищем существующую запись
			var ua models.UserAchievement
			err := db.Where("user_id = ? AND achievement_id = ?", user.ID, ach.ID).First(&ua).Error
			if err == gorm.ErrRecordNotFound {
				// Создаем только если нет записи
				ua = models.UserAchievement{
					UserID:        user.ID,
					AchievementID: ach.ID,
					Progress:      progress,
					Unlocked:      progress >= 1,
				}
				db.Create(&ua)
			}
			// Если запись есть, не трогаем, чтобы не писать в read-only
		}
	}
}

func CreateAchievement(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		Key         string          `json:"key" binding:"required"`
		Title       string          `json:"title" binding:"required"`
		Description string          `json:"description"`
		IconURL     string          `json:"icon_url"`
		Condition   json.RawMessage `json:"condition"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка, есть ли уже ачивка с таким key
	var exist models.Achievement
	if err := db.Where("key = ?", input.Key).First(&exist).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "achievement already exists"})
		return
	}

	ach := models.Achievement{
		Key:         input.Key,
		Title:       input.Title,
		Description: input.Description,
		IconURL:     input.IconURL,
		Condition:   input.Condition,
	}

	if err := db.Create(&ach).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create achievement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"achievement": ach})
}


// Пример функции calculateProgress
func calculateProgress(db *gorm.DB, userID uint, ach models.Achievement) float64 {
	var cond map[string]interface{}
	json.Unmarshal(ach.Condition, &cond)

	switch cond["type"] {
	case "story_count":
		var count int64
		db.Model(&models.Story{}).Where("user_id = ?", userID).Count(&count)
		target := int64(cond["value"].(float64))
		return float64(count) / float64(target)
	}
	return 0
}
