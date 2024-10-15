package main

import (
	"log"
	"logi/internal/api"
	"logi/internal/handlers"
	"logi/internal/repositories"
	"logi/internal/services"
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

	// Initialize repositories
	userRepo := repositories.NewUserRepository(dbClient)
	bookingRepo := repositories.NewBookingRepository(dbClient)
	driverRepo := repositories.NewDriverRepository(dbClient)

	// Initialize services
	pricingService := services.NewPricingService(bookingRepo, driverRepo)
	userService := services.NewUserService(userRepo, authService)
	bookingService := services.NewBookingService(bookingRepo, driverRepo, pricingService)
	driverService := services.NewDriverService(driverRepo, bookingRepo, authService)

	// Initialize handlers and pass the auth service where needed
	userHandler := handlers.NewUserHandler(userService, authService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	driverHandler := handlers.NewDriverHandler(driverService, authService)
	websocketHandler := handlers.NewWebSocketHandler()

	// Start the WebSocket server (to broadcast messages to clients)
	websocket.StartWebSocketServer() // Add this line to start WebSocket server

	// Initialize router
	router := api.SetupRouter(userHandler, bookingHandler, driverHandler, websocketHandler, authService)

	// Start Scheduler
	scheduler.StartScheduler(bookingService)

	// Start the server
	if err := router.Run(config.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
