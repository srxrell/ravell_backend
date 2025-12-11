package handlers

import (
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
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
