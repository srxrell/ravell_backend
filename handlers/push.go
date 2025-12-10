package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Переменные из .env
var (
	OneSignalAppID  = os.Getenv("ONESIGNAL_APP_ID")
	OneSignalAPIKey = os.Getenv("ONESIGNAL_REST_KEY")
)

// SendPushNotificationHandler — endpoint для отправки пушей вручную
func SendPushNotificationHandler(c *gin.Context) {
	var req struct {
		PlayerIDs []string `json:"player_ids"`
		Title     string   `json:"title"`
		Message   string   `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(req.PlayerIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No player IDs provided"})
		return
	}

	if err := sendPush(req.PlayerIDs, req.Title, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Push sent successfully"})
}

// sendPush — универсальная функция для пушей
func sendPush(playerIDs []string, title, message string) error {
	payload := map[string]interface{}{
		"app_id":             OneSignalAppID,
		"include_player_ids": playerIDs,
		"headings":           map[string]string{"en": title},
		"contents":           map[string]string{"en": message},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[sendPush] marshal error: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", "https://onesignal.com/api/v1/notifications", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[sendPush] new request error: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+OneSignalAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[sendPush] request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[sendPush] OneSignal error %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("OneSignal error %d: %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[sendPush] Push sent to %d devices", len(playerIDs))
	return nil
}
