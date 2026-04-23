package utils

import (
	"logi/pkg/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware(authService *auth.AuthService, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Warn(c.Request.Context(), "authorization header missing", "method", c.Request.Method, "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			Warn(c.Request.Context(), "invalid authorization header", "method", c.Request.Method, "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		token := parts[1]
		userID, role, err := authService.ValidateJWT(token)
		if err != nil {
			Warn(c.Request.Context(), "jwt validation failed", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("role", role)

		// Check if the role is allowed
		if len(requiredRoles) > 0 {
			roleAllowed := false
			for _, r := range requiredRoles {
				if role == r {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed {
				Warn(c.Request.Context(), "insufficient permissions", "user_id", userID, "role", role, "required_roles", requiredRoles)
				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
