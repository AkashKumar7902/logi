// internal/messaging/websocket_client.go
package messaging

import "logi/internal/utils"

type WebSocketClient struct {
    Hub *utils.WebSocketHub
}

func NewWebSocketClient(hub *utils.WebSocketHub) *WebSocketClient {
    return &WebSocketClient{Hub: hub}
}

func (w *WebSocketClient) Publish(userID string, messageType string, payload interface{}) error {
    message := utils.WebSocketMessage{
        UserID:  userID,
        Type:    messageType,
        Payload: payload,
    }
    w.Hub.Broadcast(message)
    return nil
}
