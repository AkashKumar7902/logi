package websocket

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrBroadcastQueueFull = errors.New("websocket broadcast queue is full")

type WebSocketHub struct {
	userClients  map[string]map[*websocket.Conn]struct{}
	adminClients map[string]map[*websocket.Conn]struct{}
	clientsMu    sync.RWMutex
	broadcast    chan WebSocketMessage
}

type WebSocketMessage struct {
	UserID  string      `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type clientConnection struct {
	userID string
	role   string
	conn   *websocket.Conn
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		userClients:  make(map[string]map[*websocket.Conn]struct{}),
		adminClients: make(map[string]map[*websocket.Conn]struct{}),
		broadcast:    make(chan WebSocketMessage, 256),
	}
}

func (hub *WebSocketHub) Run() {
	for msg := range hub.broadcast {
		adminSnapshot, userSnapshot := hub.snapshotConnections(msg.UserID)

		for _, client := range adminSnapshot {
			if err := client.conn.WriteJSON(msg); err != nil {
				hub.UnregisterClient(client.userID, client.role, client.conn)
			}
		}

		for _, client := range userSnapshot {
			if err := client.conn.WriteJSON(msg); err != nil {
				hub.UnregisterClient(client.userID, client.role, client.conn)
			}
		}
	}
}

func (hub *WebSocketHub) snapshotConnections(userID string) ([]clientConnection, []clientConnection) {
	hub.clientsMu.RLock()
	defer hub.clientsMu.RUnlock()

	adminSnapshot := make([]clientConnection, 0, len(hub.adminClients))
	for id, conns := range hub.adminClients {
		for conn := range conns {
			adminSnapshot = append(adminSnapshot, clientConnection{userID: id, role: "admin", conn: conn})
		}
	}

	userSnapshot := make([]clientConnection, 0)
	if userID != "" {
		for conn := range hub.userClients[userID] {
			userSnapshot = append(userSnapshot, clientConnection{userID: userID, role: "user", conn: conn})
		}
	}

	return adminSnapshot, userSnapshot
}

func (hub *WebSocketHub) RegisterClient(userID string, role string, conn *websocket.Conn) {
	hub.clientsMu.Lock()
	defer hub.clientsMu.Unlock()

	if role == "admin" {
		if hub.adminClients[userID] == nil {
			hub.adminClients[userID] = make(map[*websocket.Conn]struct{})
		}
		hub.adminClients[userID][conn] = struct{}{}
	} else {
		if hub.userClients[userID] == nil {
			hub.userClients[userID] = make(map[*websocket.Conn]struct{})
		}
		hub.userClients[userID][conn] = struct{}{}
	}
}

func (hub *WebSocketHub) UnregisterClient(userID string, role string, conn *websocket.Conn) {
	hub.clientsMu.Lock()
	defer hub.clientsMu.Unlock()

	if role == "admin" {
		if conns, ok := hub.adminClients[userID]; ok {
			conn.Close()
			delete(conns, conn)
			if len(conns) == 0 {
				delete(hub.adminClients, userID)
			}
		}
	} else {
		if conns, ok := hub.userClients[userID]; ok {
			conn.Close()
			delete(conns, conn)
			if len(conns) == 0 {
				delete(hub.userClients, userID)
			}
		}
	}
}

func (hub *WebSocketHub) Broadcast(msg WebSocketMessage) error {
	select {
	case hub.broadcast <- msg:
		return nil
	default:
		return ErrBroadcastQueueFull
	}
}
