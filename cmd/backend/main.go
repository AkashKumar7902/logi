package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"logi/internal/api"
	"logi/internal/handlers"
	"logi/internal/messaging"
	"logi/internal/repositories"
	"logi/internal/services"
	"logi/internal/services/distance"
	"logi/internal/utils"
	"logi/pkg/auth"
	"logi/pkg/scheduler"
	"logi/pkg/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	config, err := utils.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	utils.SetDBOperationTimeout(time.Duration(config.DBOperationTimeoutSeconds) * time.Second)

	dbClient, err := utils.ConnectDB(config.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer utils.DisconnectDB(dbClient)

	authService := auth.NewAuthService(config.JWTSecret, config.JWTExpirationHours)

	wsHub := websocket.NewWebSocketHub()
	go wsHub.Run()

	var messagingClient messaging.MessagingClient
	if config.MessagingType == "nats" {
		natsClient, err := messaging.NewNATSClient(config.NATSURL)
		if err != nil {
			log.Fatalf("Failed to connect to NATS: %v", err)
		}
		defer natsClient.Conn.Close()
		messagingClient = natsClient
	} else {
		messagingClient = messaging.NewWebSocketClient(wsHub)
	}

	userRepo := repositories.NewUserRepository(dbClient)
	bookingRepo := repositories.NewBookingRepository(dbClient)
	driverRepo := repositories.NewDriverRepository(dbClient)
	adminRepo := repositories.NewAdminRepository(dbClient)
	vehicleRepo := repositories.NewVehicleRepository(dbClient)

	var distanceCalc distance.DistanceCalculator
	switch config.DistanceCalculatorType {
	case "google_maps":
		distanceCalc = distance.NewGoogleMapsCalculator(config.GoogleMapsAPIKey)
	default:
		distanceCalc = distance.NewHaversineCalculator()
	}

	pricingService := services.NewPricingService(bookingRepo, driverRepo, distanceCalc)
	userService := services.NewUserService(userRepo, bookingRepo, driverRepo, authService)
	bookingService := services.NewBookingService(bookingRepo, driverRepo, pricingService, messagingClient)
	driverService := services.NewDriverService(driverRepo, bookingRepo, userRepo, *bookingService, authService, messagingClient)
	adminService := services.NewAdminService(adminRepo, authService, userRepo, driverRepo, bookingRepo)
	vehicleService := services.NewVehicleService(vehicleRepo)

	userHandler := handlers.NewUserHandler(userService, authService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	driverHandler := handlers.NewDriverHandler(driverService, authService)
	adminHandler := handlers.NewAdminHandler(adminService, authService, userService, driverService, bookingService, vehicleService)
	testHandler := handlers.NewTestHandler(messagingClient)

	router := api.SetupRouter(userHandler, bookingHandler, driverHandler, adminHandler, authService, wsHub, testHandler, config)

	bookingScheduler := scheduler.StartScheduler(bookingService)

	server := &http.Server{
		Addr:         config.ServerAddress,
		Handler:      router,
		ReadTimeout:  time.Duration(config.HTTPReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(config.HTTPWriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(config.HTTPIdleTimeoutSeconds) * time.Second,
	}

	go func() {
		log.Printf("Server listening on %s", config.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignals
	log.Println("Shutdown signal received")

	cronCtx := bookingScheduler.Stop()
	<-cronCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(config.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
