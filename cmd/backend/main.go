package main

import (
	"log"
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
)

func main() {
	// Load configuration
	config, err := utils.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB client
	dbClient, err := utils.ConnectDB(config.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer utils.DisconnectDB(dbClient)

	// Initialize the AuthService with the JWT secret
	authService := auth.NewAuthService(config.JWTSecret)

	wsHub := websocket.NewWebSocketHub()
	go wsHub.Run()

    // Initialize MessagingClient based on configuration
    var messagingClient messaging.MessagingClient
    if config.MessagingType == "nats" {
        natsClient, err := messaging.NewNATSClient(config.NATSURL)
        if err != nil {
            log.Fatalf("Failed to connect to NATS: %v", err)
        }
        defer natsClient.Conn.Close()
        messagingClient = natsClient
    } else {
        // Default to WebSocketClient
        messagingClient = messaging.NewWebSocketClient(wsHub)
    }

	// Initialize repositories
	userRepo := repositories.NewUserRepository(dbClient)
	bookingRepo := repositories.NewBookingRepository(dbClient)
	driverRepo := repositories.NewDriverRepository(dbClient)
	adminRepo := repositories.NewAdminRepository(dbClient)
	vehicleRepo := repositories.NewVehicleRepository(dbClient)


	// Initialize DistanceCalculator based on configuration
	var distanceCalc distance.DistanceCalculator
	switch config.DistanceCalculatorType {
	case "google_maps":
		distanceCalc = distance.NewGoogleMapsCalculator(config.GoogleMapsAPIKey)
	default:
		distanceCalc = distance.NewHaversineCalculator()
	}	

	// Initialize services
	pricingService := services.NewPricingService(bookingRepo, driverRepo, distanceCalc)
	userService := services.NewUserService(userRepo, bookingRepo, driverRepo, authService)
    bookingService := services.NewBookingService(bookingRepo, driverRepo, pricingService, messagingClient)
    driverService := services.NewDriverService(driverRepo, bookingRepo, userRepo, *bookingService, authService, messagingClient)
	adminService := services.NewAdminService(adminRepo, authService)
	vehicleService := services.NewVehicleService(vehicleRepo)

	// Initialize handlers and pass the auth service where needed
	userHandler := handlers.NewUserHandler(userService, authService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	driverHandler := handlers.NewDriverHandler(driverService, authService)
    adminHandler := handlers.NewAdminHandler(adminService, authService, userService, driverService, bookingService, vehicleService)

	// Initialize TestHandler
	testHandler := handlers.NewTestHandler(messagingClient)

	// Initialize router
	router := api.SetupRouter(userHandler, bookingHandler, driverHandler, adminHandler, authService, wsHub, testHandler)

	// Start Scheduler
	scheduler.StartScheduler(bookingService)

	// Start the server
	if err := router.Run(config.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
