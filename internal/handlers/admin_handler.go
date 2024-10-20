package handlers

import (
	"logi/internal/models"
	"logi/internal/services"
	"logi/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	Service        *services.AdminService
	AuthService    *auth.AuthService
	UserService    *services.UserService
	DriverService  *services.DriverService
	BookingService *services.BookingService
	VehicleService *services.VehicleService
}

func NewAdminHandler(service *services.AdminService, authService *auth.AuthService, userService *services.UserService, driverService *services.DriverService, bookingService *services.BookingService, vehicleService *services.VehicleService) *AdminHandler {
	return &AdminHandler{
		Service:        service,
		AuthService:    authService,
		UserService:    userService,
		DriverService:  driverService,
		BookingService: bookingService,
		VehicleService: vehicleService,
	}
}

func (h *AdminHandler) Register(c *gin.Context) {
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	admin := &models.Admin{
		Name:  payload.Name,
		Email: payload.Email,
	}

	err := h.Service.Register(admin, payload.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Admin registered successfully"})
}

func (h *AdminHandler) Login(c *gin.Context) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	admin, err := h.Service.Login(payload.Email, payload.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := h.AuthService.GenerateJWT(admin.ID, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Fleet Management Endpoints
func (h *AdminHandler) GetAllDrivers(c *gin.Context) {
	drivers, err := h.DriverService.GetAllDrivers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drivers"})
		return
	}
	c.JSON(http.StatusOK, drivers)
}

func (h *AdminHandler) GetDriver(c *gin.Context) {
	driverID := c.Param("driverID")
	driver, err := h.DriverService.GetDriverByID(driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}
	c.JSON(http.StatusOK, driver)
}

func (h *AdminHandler) UpdateDriver(c *gin.Context) {
	driverID := c.Param("driverID")
	var driver models.Driver
	if err := c.BindJSON(&driver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
    oldDriver, err := h.DriverService.GetDriverByID(driverID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
        return
    }
	if driver.VehicleID != "" {
    	oldDriver.VehicleID = driver.VehicleID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A vehicle has already been assigned to this driver"})
		return
	}
    
	err = h.DriverService.UpdateDriver(oldDriver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Driver updated successfully"})
}

// Vehicle Management Endpoints

func (h *AdminHandler) CreateVehicle(c *gin.Context) {
	var vehicle models.Vehicle
	if err := c.BindJSON(&vehicle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.VehicleService.CreateVehicle(&vehicle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vehicle"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Vehicle created successfully", "vehicle": vehicle})
}

func (h *AdminHandler) UpdateVehicle(c *gin.Context) {
	vehicleID := c.Param("vehicleID")
	var vehicle models.Vehicle
	if err := c.BindJSON(&vehicle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	vehicle.ID = vehicleID
	err := h.VehicleService.UpdateVehicle(&vehicle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vehicle"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Vehicle updated successfully", "vehicle": vehicle})
}

func (h *AdminHandler) DeleteVehicle(c *gin.Context) {
	vehicleID := c.Param("vehicleID")
	err := h.VehicleService.DeleteVehicle(vehicleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vehicle"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Vehicle deleted successfully"})
}

func (h *AdminHandler) GetVehicle(c *gin.Context) {
	vehicleID := c.Param("vehicleID")
	vehicle, err := h.VehicleService.GetVehicleByID(vehicleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

func (h *AdminHandler) GetAllVehicles(c *gin.Context) {
	vehicles, err := h.VehicleService.GetAllVehicles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vehicles"})
		return
	}
	c.JSON(http.StatusOK, vehicles)
}

func (h *AdminHandler) GetStatistics(c *gin.Context) {
	stats, err := h.Service.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
