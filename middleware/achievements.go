package middleware

import (
	"fmt"
	"go_stories_api/handlers"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// AchievementMiddleware универсальный для любых роутов.
// Проверяет параметр `id` (userID) и апдейтит прогресс ачивок.
func AchievementMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем userID из параметров
		idParam := c.Param("id")
		if idParam == "" {
			c.Next()
			return
		}

		var userID uint
		if _, err := fmt.Sscan(idParam, &userID); err != nil {
			c.Next()
			return
		}

		// Пример триггеров ачивок
		// 1) За заход на профиль
		handlers.UpdateAchievementProgress(db, userID, "view_profile", 1)

		// 2) За лайк поста
		if c.FullPath() == "/stories/:id/like" && c.Request.Method == "POST" {
			handlers.UpdateAchievementProgress(db, userID, "first_like", 1)
		}

		// 3) За написание первой истории
		if c.FullPath() == "/stories/" && c.Request.Method == "POST" {
			handlers.UpdateAchievementProgress(db, userID, "first_story", 1)
		}

		// Можно добавить любые другие события по пути, методу или body запроса

		c.Next()
	}
}
