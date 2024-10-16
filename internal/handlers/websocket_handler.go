package handlers

import (
	"logi/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	websk "github.com/gorilla/websocket"
    "logi/pkg/websocket"

)

var upgrader = websk.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for simplicity; adjust as needed
    },
}

func ServeWs(authService *auth.AuthService, hub *websocket.WebSocketHub, c *gin.Context) {
    tokenString := c.Query("token")
    if tokenString == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token query parameter missing"})
        return
    }

    // Validate JWT token
    userID, role, err := authService.ValidateJWT(tokenString)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    hub.RegisterClient(userID, role, conn)

    // Handle incoming messages if needed
    go func() {
        defer hub.UnregisterClient(userID, role)
        for {
            _, _, err := conn.ReadMessage()
            if err != nil {
                break
            }
        }
    }()
}
