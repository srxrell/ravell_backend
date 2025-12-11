package wsservice

import (
	"sync"

	"github.com/gorilla/websocket"
)

// map[userID] -> список соединений
var connections = struct {
	sync.RWMutex
	clients map[uint][]*websocket.Conn
}{clients: make(map[uint][]*websocket.Conn)}

// Добавить соединение
func AddConnection(userID uint, conn *websocket.Conn) {
	connections.Lock()
	defer connections.Unlock()
	connections.clients[userID] = append(connections.clients[userID], conn)
}

// Удалить соединение
func RemoveConnection(userID uint, conn *websocket.Conn) {
	connections.Lock()
	defer connections.Unlock()
	conns := connections.clients[userID]
	for i, c := range conns {
		if c == conn {
			connections.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

// Отправить уведомление конкретному пользователю
func SendNotification(userID uint, message interface{}) {
	connections.RLock()
	defer connections.RUnlock()
	conns := connections.clients[userID]
	for _, c := range conns {
		c.WriteJSON(message)
	}
}
