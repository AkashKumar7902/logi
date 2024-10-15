package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var mutex = &sync.Mutex{}

// Message defines the structure of messages
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}

		broadcast <- msg
	}
}

// StartWebSocketServer starts a background goroutine to handle message broadcasting
func StartWebSocketServer() {
	go func() {
		for {
			msg := <-broadcast // Receive a message from the broadcast channel

			// Send the message to all connected clients
			mutex.Lock()
			for client := range clients {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Println("WebSocket write error:", err)
					client.Close()          // Close client connection on error
					delete(clients, client) // Remove from active clients
				}
			}
			mutex.Unlock()
		}
	}()
}
