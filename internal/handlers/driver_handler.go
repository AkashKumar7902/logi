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
		Service: service,
		AuthService: authService,
	}
}

func (h *DriverHandler) Register(c *gin.Context) {
	var payload struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		VehicleType string `json:"vehicle_type"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	driver := &models.Driver{
		Name:        payload.Name,
		Email:       payload.Email,
		VehicleType: payload.VehicleType,
	}

	err := h.Service.Register(driver, payload.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Driver registered successfully"})
}

func (h *DriverHandler) Login(c *gin.Context) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	driver, err := h.Service.Login(payload.Email, payload.Password)
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
	var payload struct {
		Status string `json:"status"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	driverID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.Service.UpdateStatus(driverID.(string), payload.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

// UpdateStatus allows drivers to update their booking status (e.g., En Route, Goods Collected, etc.)
func (h *DriverHandler) UpdateBookingStatus(c *gin.Context) {
    var req struct {
        BookingID string `json:"booking_id"`
        Status    string `json:"status"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    driverID, exists := c.Get("userID") // Assuming driverID is stored in JWT token
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Update the booking status through the service layer
    err := h.Service.UpdateBookingStatus(driverID.(string), req.BookingID, req.Status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Booking status updated successfully"})
}

func (h *DriverHandler) UpdateLocation(c *gin.Context) {
    var payload struct {
        Latitude  float64 `json:"latitude"`
        Longitude float64 `json:"longitude"`
    }

    if err := c.BindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    driverID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    err := h.Service.UpdateLocation(driverID.(string), payload.Latitude, payload.Longitude)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Location updated"})
}
