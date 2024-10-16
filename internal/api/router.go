package api

import (
	"logi/internal/handlers"
	"logi/internal/utils"
	"logi/pkg/auth"
	"logi/pkg/websocket"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	userHandler *handlers.UserHandler,
	bookingHandler *handlers.BookingHandler,
	driverHandler *handlers.DriverHandler,
	adminHandler *handlers.AdminHandler,
	authService *auth.AuthService,
	wsHub *websocket.WebSocketHub,
) *gin.Engine {
	router := gin.Default()

	// Public routes
	router.POST("/users/register", userHandler.Register)
	router.POST("/users/login", userHandler.Login)
	router.POST("/drivers/register", driverHandler.Register)
	router.POST("/drivers/login", driverHandler.Login)
	router.POST("/admins/register", adminHandler.Register)
	router.POST("/admins/login", adminHandler.Login)

	router.GET("/ws", func(c *gin.Context) {
		handlers.ServeWs(authService, wsHub, c)
	})

	// Protected routes with JWT middleware
	userProtected := router.Group("/", utils.JWTAuthMiddleware(authService, "user"))
	{
		userProtected.POST("/bookings", bookingHandler.CreateBooking)
		userProtected.POST("/bookings/estimate", bookingHandler.GetPriceEstimate)
	}

	driverProtected := router.Group("/drivers", utils.JWTAuthMiddleware(authService, "driver"))
	{
		driverProtected.POST("/status", driverHandler.UpdateStatus)
		driverProtected.POST("/booking-status", driverHandler.UpdateBookingStatus)
		driverProtected.POST("/update-location", driverHandler.UpdateLocation)
	}

	adminProtected := router.Group("/admin", utils.JWTAuthMiddleware(authService, "admin"))
	{
		adminProtected.GET("/drivers", adminHandler.GetAllDrivers)
		adminProtected.GET("/drivers/:driverID", adminHandler.GetDriver)
		adminProtected.PUT("/drivers/:driverID", adminHandler.UpdateDriver)
		adminProtected.GET("/analytics", adminHandler.GetAnalytics)
	}

	return router
}
