package api

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"logi/internal/handlers"
	"logi/internal/utils"
	"logi/pkg/auth"
	"logi/pkg/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const adminBootstrapSecretHeader = "X-Admin-Bootstrap-Secret"

type ReadinessCheck func(context.Context) error

func corsMiddleware(cfg *utils.Config) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", utils.RequestIDHeader},
		ExposeHeaders:    []string{utils.RequestIDHeader},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	allowAll := false
	for _, origin := range cfg.AllowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
	}

	if allowAll {
		config.AllowAllOrigins = true
	} else {
		config.AllowOrigins = cfg.AllowedOrigins
	}

	return cors.New(config)
}

func adminBootstrapMiddleware(cfg *utils.Config) gin.HandlerFunc {
	expectedSecret := strings.TrimSpace(cfg.AdminBootstrapSecret)

	return func(c *gin.Context) {
		providedSecret := strings.TrimSpace(c.GetHeader(adminBootstrapSecretHeader))
		if providedSecret == "" || subtle.ConstantTimeCompare([]byte(providedSecret), []byte(expectedSecret)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid bootstrap secret"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func SetupRouter(
	userHandler *handlers.UserHandler,
	bookingHandler *handlers.BookingHandler,
	driverHandler *handlers.DriverHandler,
	adminHandler *handlers.AdminHandler,
	authService *auth.AuthService,
	wsHub *websocket.WebSocketHub,
	testHandler *handlers.TestHandler,
	cfg *utils.Config,
	readinessChecks ...ReadinessCheck,
) *gin.Engine {
	router := gin.New()
	router.Use(utils.RequestIDMiddleware())
	router.Use(utils.RequestLoggingMiddleware())
	router.Use(utils.RecoveryMiddleware())
	router.Use(corsMiddleware(cfg))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/readyz", func(c *gin.Context) {
		for _, check := range readinessChecks {
			if err := check(c.Request.Context()); err != nil {
				utils.Warn(c.Request.Context(), "readiness check failed", "error", err)
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "not_ready",
					"error":  "dependency check failed",
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Public routes
	router.POST("/users/register", userHandler.Register)
	router.POST("/users/login", userHandler.Login)
	router.POST("/drivers/register", driverHandler.Register)
	router.POST("/drivers/login", driverHandler.Login)
	router.POST("/admins/login", adminHandler.Login)
	if cfg.EnableAdminBootstrap {
		router.POST("/internal/bootstrap/admin", adminBootstrapMiddleware(cfg), adminHandler.RegisterBootstrap)
	}

	router.GET("/ws", func(c *gin.Context) {
		handlers.ServeWs(authService, wsHub, cfg.AllowedOriginsSet(), c)
	})

	if cfg.EnableTestRoutes {
		router.GET("/test", utils.JWTAuthMiddleware(authService, "admin"), testHandler.PublishTestMessages)
	}

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
