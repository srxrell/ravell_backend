package handlers

import (
	"go_stories_api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SavePlayerID(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)
    userID := c.MustGet("user_id").(uint)

    var req struct {
        PlayerID string `json:"player_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    device := models.UserDevice{
        UserID:   userID,
        PlayerID: req.PlayerID,
    }

    if err := db.Create(&device).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save playerId"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "PlayerId saved"})
}
