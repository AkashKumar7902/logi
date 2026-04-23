package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	userClients  map[string]*websocket.Conn
	adminClients map[string]*websocket.Conn
	clientsMu    sync.RWMutex
	broadcast    chan WebSocketMessage
}

type WebSocketMessage struct {
	UserID  string      `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		userClients:  make(map[string]*websocket.Conn),
		adminClients: make(map[string]*websocket.Conn),
		broadcast:    make(chan WebSocketMessage, 256),
	}
}

func (hub *WebSocketHub) Run() {
	for msg := range hub.broadcast {
		adminSnapshot, userConn := hub.snapshotConnections(msg.UserID)

		for adminID, conn := range adminSnapshot {
			if err := conn.WriteJSON(msg); err != nil {
				hub.UnregisterClient(adminID, "admin")
			}
		}

		if userConn != nil {
			if err := userConn.WriteJSON(msg); err != nil {
				hub.UnregisterClient(msg.UserID, "user")
			}
		}
	}
}

func (hub *WebSocketHub) snapshotConnections(userID string) (map[string]*websocket.Conn, *websocket.Conn) {
	hub.clientsMu.RLock()
	defer hub.clientsMu.RUnlock()

	adminSnapshot := make(map[string]*websocket.Conn, len(hub.adminClients))
	for id, conn := range hub.adminClients {
		adminSnapshot[id] = conn
	}

	var userConn *websocket.Conn
	if userID != "" {
		userConn = hub.userClients[userID]
	}

	return adminSnapshot, userConn
}

func (hub *WebSocketHub) RegisterClient(userID string, role string, conn *websocket.Conn) {
	hub.clientsMu.Lock()
	if role == "admin" {
		hub.adminClients[userID] = conn
	} else {
		hub.userClients[userID] = conn
	}
	hub.clientsMu.Unlock()
}

func (hub *WebSocketHub) UnregisterClient(userID string, role string) {
	hub.clientsMu.Lock()
	if role == "admin" {
		if conn, ok := hub.adminClients[userID]; ok {
			conn.Close()
			delete(hub.adminClients, userID)
		}
	} else {
		if conn, ok := hub.userClients[userID]; ok {
			conn.Close()
			delete(hub.userClients, userID)
		}
	}
	hub.clientsMu.Unlock()
}

func (hub *WebSocketHub) Broadcast(msg WebSocketMessage) {
	hub.broadcast <- msg
}
