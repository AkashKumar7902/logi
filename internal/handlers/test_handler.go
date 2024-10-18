// internal/handlers/test_handler.go
package handlers

import (
	"net/http"

	"logi/internal/messaging"
	"logi/internal/utils"

	"github.com/gin-gonic/gin"
)

type TestHandler struct {
	MessagingClient messaging.MessagingClient
}

func NewTestHandler(messagingClient messaging.MessagingClient) *TestHandler {
	return &TestHandler{
		MessagingClient: messagingClient,
	}
}

// PublishTestMessages publishes different types of test messages
func (h *TestHandler) PublishTestMessages(c *gin.Context) {
	userID := c.GetString("userID")
	utils.Logger.Println("userid", userID)

	// Example: Publish a status update message
	statusUpdate := map[string]interface{}{
		"type":       "status_update",
		"booking_id": "booking123",
		"status":     "In Transit",
	}

	err := h.MessagingClient.Publish(userID, "status_update", statusUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish status update"})
		return
	}

	// Example: Publish a driver location update message
	driverLocation := map[string]interface{}{
		"type":       "driver_location",
		"booking_id": "booking123",
		"latitude":   37.7749,
		"longitude":  -122.4194,
	}

	err = h.MessagingClient.Publish("user123", "driver_location", driverLocation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish driver location"})
		return
	}

	// Example: Publish a booking accepted message
	bookingAccepted := map[string]interface{}{
		"type":       "booking_accepted",
		"booking_id": "booking123",
		"driver_id":  "driver456",
	}

	err = h.MessagingClient.Publish("user123", "booking_accepted", bookingAccepted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish booking accepted"})
		return
	}

	err = h.MessagingClient.Publish("b8fe009c-7cf4-435f-9161-59a0f954c5c4", "new_booking_request", map[string]interface{}{"booking_id": "booking123"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish new booking request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test messages published successfully"})
}
