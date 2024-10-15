package api

import (
    "logi/internal/handlers"
    "logi/internal/utils"
    "logi/pkg/auth"

    "github.com/gin-gonic/gin"
)

func SetupRouter(userHandler *handlers.UserHandler, bookingHandler *handlers.BookingHandler, driverHandler *handlers.DriverHandler, websocketHandler *handlers.WebSocketHandler, authService *auth.AuthService) *gin.Engine {
    router := gin.Default()

    // Public routes
    router.POST("/users/register", userHandler.Register)
    router.POST("/users/login", userHandler.Login)
    router.POST("/drivers/register", driverHandler.Register)
    router.POST("/drivers/login", driverHandler.Login)

    // WebSocket endpoint
    router.GET("/ws", websocketHandler.HandleWebSocket)

    // Protected routes with JWT middleware
    userProtected := router.Group("/", utils.JWTAuthMiddleware(authService))
    {
        userProtected.POST("/bookings", bookingHandler.CreateBooking)
    }

    driverProtected := router.Group("/drivers", utils.JWTAuthMiddleware(authService))
    {
        driverProtected.POST("/status", driverHandler.UpdateStatus)
    }

    return router
}
