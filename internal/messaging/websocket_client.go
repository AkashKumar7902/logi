// internal/messaging/websocket_client.go
package messaging

import "logi/pkg/websocket"

type WebSocketClient struct {
    Hub *websocket.WebSocketHub
}

func NewWebSocketClient(hub *websocket.WebSocketHub) *WebSocketClient {
    return &WebSocketClient{Hub: hub}
}

func (w *WebSocketClient) Publish(userID string, messageType string, payload interface{}) error {
    message := websocket.WebSocketMessage{
        UserID:  userID,
        Type:    messageType,
        Payload: payload,
    }
    w.Hub.Broadcast(message)
    return nil
}
