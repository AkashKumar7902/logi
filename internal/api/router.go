package api

import (
	"logi/internal/handlers"
	"logi/internal/utils"
	"logi/pkg/auth"
	"logi/pkg/websocket"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Default() gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	config.AllowAllOrigins = true
	return cors.New(config)
}

func SetupRouter(
	userHandler *handlers.UserHandler,
	bookingHandler *handlers.BookingHandler,
	driverHandler *handlers.DriverHandler,
	adminHandler *handlers.AdminHandler,
	authService *auth.AuthService,
	wsHub *websocket.WebSocketHub,
	testHandler *handlers.TestHandler,
) *gin.Engine {
	router := gin.Default()
	router.Use(Default())

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
	
	router.GET("/test", testHandler.PublishTestMessages)

	// Protected routes with JWT middleware
	userProtected := router.Group("/", utils.JWTAuthMiddleware(authService, "user"))
	{
		userProtected.GET("/active-booking", userHandler.GetActiveBooking)
		userProtected.GET("/bookings/:bookingID/driver", userHandler.GetDriverForBooking)
		userProtected.POST("/bookings", bookingHandler.CreateBooking)
		userProtected.POST("/bookings/estimate", bookingHandler.GetPriceEstimate)
	}

	driverProtected := router.Group("/drivers", utils.JWTAuthMiddleware(authService, "driver"))
	{
		driverProtected.GET("/active-bookings", driverHandler.GetActiveBookings)
		driverProtected.GET("/bookings/:bookingID/user", driverHandler.GetUserForBooking)
		driverProtected.GET("/bookings/:bookingID", driverHandler.GetBooking)
		driverProtected.GET("/me", driverHandler.GetDriverInfo)
		driverProtected.POST("/status", driverHandler.UpdateStatus)
		driverProtected.POST("/booking-status", driverHandler.UpdateBookingStatus)
		driverProtected.POST("/update-location", driverHandler.UpdateLocation)
		driverProtected.GET("/pending-bookings", driverHandler.GetPendingBookings)
		driverProtected.POST("/respond-booking", driverHandler.RespondToBooking)
	}

	adminProtected := router.Group("/admin", utils.JWTAuthMiddleware(authService, "admin"))
	{
		adminProtected.GET("/drivers", adminHandler.GetAllDrivers)
		adminProtected.GET("/drivers/:driverID", adminHandler.GetDriver)
		adminProtected.PUT("/drivers/:driverID", adminHandler.UpdateDriver)

		adminProtected.GET("/statistics", adminHandler.GetStatistics)

		// Vehicle management routes
		adminProtected.POST("/vehicles", adminHandler.CreateVehicle)
		adminProtected.GET("/vehicles", adminHandler.GetAllVehicles)
		adminProtected.GET("/vehicles/:vehicleID", adminHandler.GetVehicle)
		adminProtected.PUT("/vehicles/:vehicleID", adminHandler.UpdateVehicle)
		adminProtected.DELETE("/vehicles/:vehicleID", adminHandler.DeleteVehicle)

	}

	return router
}
