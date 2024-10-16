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
}

func NewAdminHandler(service *services.AdminService, authService *auth.AuthService, userService *services.UserService, driverService *services.DriverService, bookingService *services.BookingService) *AdminHandler {
    return &AdminHandler{
        Service:        service,
        AuthService:    authService,
        UserService:    userService,
        DriverService:  driverService,
        BookingService: bookingService,
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
    driver.ID = driverID
    err := h.DriverService.UpdateDriver(&driver)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Driver updated successfully"})
}

func (h *AdminHandler) GetAnalytics(c *gin.Context) {
    stats, err := h.BookingService.GetBookingStatistics()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
        return
    }
    c.JSON(http.StatusOK, stats)
}
