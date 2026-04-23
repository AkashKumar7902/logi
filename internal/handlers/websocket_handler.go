package handlers

import (
	"net/http"
	"strings"

	"logi/pkg/auth"
	"logi/pkg/websocket"

	"github.com/gin-gonic/gin"
	websk "github.com/gorilla/websocket"
)

func ServeWs(authService *auth.AuthService, hub *websocket.WebSocketHub, allowedOrigins map[string]struct{}, c *gin.Context) {
	tokenString := tokenFromRequest(c)
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing"})
		return
	}

	userID, role, err := authService.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	upgrader := websk.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				// Non-browser clients may omit Origin.
				return true
			}
			if _, ok := allowedOrigins["*"]; ok {
				return true
			}
			_, ok := allowedOrigins[origin]
			return ok
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	hub.RegisterClient(userID, role, conn)

	go func() {
		defer hub.UnregisterClient(userID, role)
		for {
			if _, _, readErr := conn.ReadMessage(); readErr != nil {
				break
			}
		}
	}()
}

func tokenFromRequest(c *gin.Context) string {
	token := strings.TrimSpace(c.Query("token"))
	if token != "" {
		return token
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
