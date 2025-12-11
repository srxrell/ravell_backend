package handlers

import (
	"encoding/json"
	"go_stories_api/wsservice"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// структура для сообщений через WS
type WSMessage struct {
	Action  string `json:"action"`   // "send_to_user"
	UserID  uint   `json:"user_id"`  // кому шлём
	Message string `json:"message"`  // текст уведомления
}

func WSHandler(c *gin.Context) {
	userID := c.GetUint("userID") // берём из JWT middleware

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade"})
		return
	}

	wsservice.AddConnection(userID, conn)
	defer func() {
		wsservice.RemoveConnection(userID, conn)
		conn.Close()
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// парсим сообщение
		var msg WSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue // игнорируем кривые сообщения
		}

		// если пришёл push другому пользователю
		if msg.Action == "send_to_user" && msg.UserID != 0 && msg.Message != "" {
			wsservice.SendNotification(msg.UserID, map[string]string{
				"type":          "follow", // или другой тип
				"from_username": "Система", 
				"message":       msg.Message,
			})
		}
	}
}
