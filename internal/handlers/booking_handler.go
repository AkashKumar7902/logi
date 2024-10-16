package handlers

import (
    "net/http"
    "logi/internal/models"
    "logi/internal/services"

    "github.com/gin-gonic/gin"
)

type BookingHandler struct {
    Service *services.BookingService
}

func NewBookingHandler(service *services.BookingService) *BookingHandler {
    return &BookingHandler{Service: service}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
    var bookingReq models.BookingRequest
    if err := c.BindJSON(&bookingReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    booking, err := h.Service.CreateBooking(userID.(string), &bookingReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) GetPriceEstimate(c *gin.Context) {
    var estimateReq models.PriceEstimateRequest

    // Bind and validate the JSON input
    if err := c.ShouldBindJSON(&estimateReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
        return
    }

    // Call the service to get the price estimate
    estimatedPrice, err := h.Service.GetPriceEstimate(&estimateReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate price estimate"})
        return
    }

    // Prepare the response
    response := models.PriceEstimateResponse{
        EstimatedPrice: estimatedPrice,
    }

    // Return the estimated price
    c.JSON(http.StatusOK, response)
}