package handlers

import (
	"logi/internal/models"
	"logi/internal/services"
	"logi/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DriverHandler struct {
	Service     *services.DriverService
	AuthService *auth.AuthService
}

func NewDriverHandler(service *services.DriverService, authService *auth.AuthService) *DriverHandler {
	return &DriverHandler{
		Service:     service,
		AuthService: authService,
	}
}

func (h *DriverHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var payload struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		VehicleType string `json:"vehicle_type"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	payload.VehicleType = models.NormalizeVehicleType(payload.VehicleType)
	if payload.VehicleType != "" && !models.IsValidVehicleType(payload.VehicleType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_type must be one of: bike, car, van"})
		return
	}

	driver := &models.Driver{
		Name:        payload.Name,
		Email:       payload.Email,
		VehicleType: payload.VehicleType,
	}

	err := h.Service.Register(ctx, driver, payload.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Driver registered successfully"})
}

func (h *DriverHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	driver, err := h.Service.Login(ctx, payload.Email, payload.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := h.AuthService.GenerateJWT(driver.ID, "driver")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *DriverHandler) UpdateStatus(c *gin.Context) {
	ctx := c.Request.Context()

	var payload struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := models.ValidateDriverStatus(payload.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driverID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.Service.UpdateStatus(ctx, driverID.(string), payload.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

// UpdateStatus allows drivers to update their booking status (e.g., En Route, Goods Collected, etc.)
func (h *DriverHandler) UpdateBookingStatus(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		BookingID string `json:"booking_id"`
		Status    string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	driverID, exists := c.Get("userID") // Assuming driverID is stored in JWT token
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Update the booking status through the service layer
	err := h.Service.UpdateBookingStatus(ctx, driverID.(string), req.BookingID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking status updated successfully"})
}

func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	ctx := c.Request.Context()

	var payload struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := models.ValidateLatitudeLongitude(payload.Latitude, payload.Longitude); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driverID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.Service.UpdateLocation(ctx, driverID.(string), payload.Latitude, payload.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location updated"})
}

func (h *DriverHandler) GetPendingBookings(c *gin.Context) {
	ctx := c.Request.Context()
	driverID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	bookings, err := h.Service.GetPendingBookings(ctx, driverID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *DriverHandler) RespondToBooking(c *gin.Context) {
	ctx := c.Request.Context()

	var payload struct {
		BookingID string `json:"booking_id"`
		Response  string `json:"response"` // "accept" or "reject"
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() + "Invalid input"})
		return
	}

	driverID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.Service.RespondToBooking(ctx, driverID.(string), payload.BookingID, payload.Response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response recorded"})
}

func (h *DriverHandler) GetActiveBookings(c *gin.Context) {
	ctx := c.Request.Context()
	driverID := c.GetString("userID")

	bookings, err := h.Service.GetActiveBookings(ctx, driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active bookings found"})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

func (h *DriverHandler) GetUserForBooking(c *gin.Context) {
	ctx := c.Request.Context()
	driverID := c.GetString("userID")
	bookingID := c.Param("bookingID")

	user, err := h.Service.GetUserForBooking(ctx, driverID, bookingID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *DriverHandler) GetDriverInfo(c *gin.Context) {
	ctx := c.Request.Context()
	driverID := c.GetString("userID")

	driver, err := h.Service.GetDriverInfo(ctx, driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve driver info"})
		return
	}
	c.JSON(http.StatusOK, driver)
}

// GetBooking retrieves a specific booking by ID for the authenticated driver.
func (h *DriverHandler) GetBooking(c *gin.Context) {
	ctx := c.Request.Context()
	bookingID := c.Param("bookingID")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking ID is required"})
		return
	}

	driverID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	booking, err := h.Service.GetBooking(ctx, driverID.(string), bookingID)
	if err != nil {
		if err.Error() == "booking not found or not assigned to the driver" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve booking"})
		}
		return
	}

	c.JSON(http.StatusOK, booking)
}
