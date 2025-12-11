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

	// просто держим соединение открытым
	for {
    _, msg, err := conn.ReadMessage()
    if err != nil {
        break
    }

    var payload map[string]interface{}
    if err := json.Unmarshal(msg, &payload); err != nil {
        continue
    }

    action, ok := payload["action"].(string)
    if !ok {
        continue
    }

    if action == "send_to_user" {
        userIDFloat, ok := payload["user_id"].(float64)
        if !ok {
            continue
        }
        targetUserID := uint(userIDFloat)
        message := payload["message"]

        wsservice.SendNotification(targetUserID, map[string]interface{}{
            "type":    "follow",
            "message": message,
        })
    }
}

}
