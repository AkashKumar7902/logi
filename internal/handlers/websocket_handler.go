package handlers

import (
    "logi/pkg/websocket"

    "github.com/gin-gonic/gin"
)

type WebSocketHandler struct{}

func NewWebSocketHandler() *WebSocketHandler {
    return &WebSocketHandler{}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
    websocket.HandleWebSocket(c.Writer, c.Request)
}
