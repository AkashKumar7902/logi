package utils

import (
    "net/http"
    "logi/pkg/auth"
    "strings"

    "github.com/gin-gonic/gin"
)

func JWTAuthMiddleware(authService *auth.AuthService, requiredRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
            c.Abort()
            return
        }

        token := parts[1]
        userID, role, err := authService.ValidateJWT(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

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
                c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
                c.Abort()
                return
            }
        }

        c.Set("userID", userID)
        c.Set("role", role)
        c.Next()
    }
}
