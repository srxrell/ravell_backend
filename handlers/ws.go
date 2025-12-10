package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Хранилище подключений
var clients = make(map[*websocket.Conn]uint) // map[conn]userID
var clientsMu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSHandler(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = userID
	clientsMu.Unlock()
	log.Printf("User %d connected via WebSocket\n", userID)

	// Ловим отключение
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
	log.Printf("User %d disconnected\n", userID)
}

// Отправка уведомления конкретному пользователю
func SendWSNotification(userID uint, title, body string) {
	message := map[string]string{
		"title": title,
		"body":  body,
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	for conn, id := range clients {
		if id == userID {
			if err := conn.WriteJSON(message); err != nil {
				log.Println("WebSocket write error:", err)
				conn.Close()
				delete(clients, conn)
			}
		}
	}
}
