package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendPushNotification(c *gin.Context) {
	var req struct {
		PlayerIDs []string `json:"player_ids"`
		Title     string   `json:"title"`
		Message   string   `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := map[string]interface{}{
		"app_id":             "13381b7f-376e-48ca-80a4-c0b5633dfde6",
		"include_player_ids": req.PlayerIDs,
		"headings":           map[string]string{"en": req.Title},
		"contents":           map[string]string{"en": req.Message},
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", "https://onesignal.com/api/v1/notifications", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic os_v2_app_cm4bw7zxnzemvafeyc2wgpp543xz7dm242ke37efjtsrwsnsffklhxhugvhiovlyjwfakg3ya4yfrjctxsohnficlneg7aaupoxlo6q") // REST API Key из OneSignal

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}
