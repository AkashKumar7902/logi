package websocket

import (
    "sync"

    "github.com/gorilla/websocket"
)

type WebSocketHub struct {
    userClients  map[string]*websocket.Conn
    adminClients map[string]*websocket.Conn
    clientsMu  sync.RWMutex
    broadcast  chan WebSocketMessage
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
        broadcast: make(chan WebSocketMessage),
    }
}

func (hub *WebSocketHub) Run() {
    for msg := range hub.broadcast {
        hub.clientsMu.RLock()

		for _, conn := range hub.adminClients {
            err := conn.WriteJSON(msg)
            if err != nil {
                // Handle error (e.g., remove the client)
                conn.Close()
                // Note: It's generally unsafe to delete while ranging; consider collecting to delete after.
            }
        }

        if conn, ok := hub.userClients[msg.UserID]; ok {
            err := conn.WriteJSON(msg)
			if err != nil {
				conn.Close()
				delete(hub.userClients, msg.UserID)
			}
        }
        hub.clientsMu.RUnlock()
    }
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
