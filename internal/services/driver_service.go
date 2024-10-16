package services

import (
	"errors"
	"logi/internal/messaging"
	"logi/internal/models"
	"logi/internal/repositories"
	"logi/internal/utils"
	"logi/pkg/auth"
	"time"

	"github.com/google/uuid"
)

type DriverService struct {
	Repo            repositories.DriverRepository
	BookingRepo     repositories.BookingRepository
	BookingService  BookingService
	AuthService     *auth.AuthService
	MessagingClient messaging.MessagingClient
}

func NewDriverService(repo repositories.DriverRepository, bookingRepo repositories.BookingRepository, bookingService BookingService,  authService *auth.AuthService, messagingClient messaging.MessagingClient) *DriverService {
	return &DriverService{
		Repo:            repo,
		AuthService:     authService,
		BookingRepo:     bookingRepo,
		BookingService:  bookingService,
		MessagingClient: messagingClient,
	}
}

func (s *DriverService) Register(driver *models.Driver, password string) error {
	existingDriver, _ := s.Repo.FindByEmail(driver.Email)
	if existingDriver != nil {
		return errors.New("driver already exists")
	}

	hashedPassword, err := s.AuthService.HashPassword(password)
	if err != nil {
		return err
	}

	driver.ID = uuid.NewString()
	driver.PasswordHash = hashedPassword
	driver.Status = "Available"
	driver.CreatedAt = time.Now()

	return s.Repo.Create(driver)
}

func (s *DriverService) Login(email, password string) (*models.Driver, error) {
	driver, err := s.Repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !s.AuthService.CheckPasswordHash(password, driver.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return driver, nil
}

// UpdateStatus updates the driver's status and notifies admins
func (s *DriverService) UpdateStatus(driverID, status string) error {
	// Update the driver's status in the repository
	err := s.Repo.UpdateStatus(driverID, status)
	if err != nil {
		return err
	}

	// Publish the status update to admins via MessagingClient
	publishErr := s.MessagingClient.Publish("", "driver_status_update", map[string]interface{}{
		"driver_id": driverID,
		"status":    status,
	})
	if publishErr != nil {
		// Log the error but do not fail the operation
		utils.Logger.Printf("Failed to publish driver status update: %v", publishErr)
	}

	return nil
}

// UpdateBookingStatus updates the status of a booking
func (s *DriverService) UpdateBookingStatus(driverID string, bookingID string, status string) error {
	// Find the booking by ID
	booking, err := s.BookingRepo.FindByID(bookingID)
	if err != nil {
		return err
	}

	// Ensure that the driver is assigned to this booking
	if booking.DriverID != driverID {
		return errors.New("driver not assigned to this booking")
	}

	// Validate status transition
	validTransitions := map[string][]string{
		"Driver Assigned":    {"En Route to Pickup"},
		"En Route to Pickup": {"Goods Collected"},
		"Goods Collected":    {"In Transit"},
		"In Transit":         {"Delivered"},
		"Delivered":          {"Completed"},
	}

	currentStatus := booking.Status
	if !isValidTransition(currentStatus, status, validTransitions) {
		return errors.New("invalid status transition")
	}

	// Update timestamps based on status
	currentTime := time.Now()
	switch status {
	case "In Transit":
		if booking.StartedAt == nil {
			booking.StartedAt = &currentTime
		}
	case "Completed":
		if booking.CompletedAt == nil {
			booking.CompletedAt = &currentTime
		}
	}

	err = s.BookingRepo.Update(booking)
	if err != nil {
		return err
	}

	// Update the booking status
	booking.Status = status
	err = s.BookingRepo.Update(booking)
	if err != nil {
		return err
	}

	// Notify user about status update
	s.MessagingClient.Publish(booking.UserID, "status_update", map[string]interface{}{
		"booking_id": booking.ID,
		"status":     status,
	})

	// If booking is marked as completed, update driver's status to Available
	if status == "Completed" {
		err = s.UpdateStatus(booking.DriverID, "Available")
		if err != nil {
			return errors.New("failed to update driver status to Available")
		}
	}

	return nil
}

func (s *DriverService) UpdateLocation(driverID string, latitude, longitude float64) error {
	location := models.Location{
        Type:        "Point",
        Coordinates: []float64{longitude, latitude},
    }
	
	err := s.Repo.UpdateLocation(driverID, location)
	if err != nil {
		return err
	}

	// Find current booking for the driver
	booking, err := s.BookingRepo.FindActiveBookingByDriverID(driverID)
	if err != nil {
		return nil // No active booking, no need to notify
	}

	// Notify user about driver's location update
	s.MessagingClient.Publish(booking.UserID, "driver_location", map[string]interface{}{
		"booking_id": booking.ID,
		"latitude":   latitude,
		"longitude":  longitude,
	})

	return nil
}

func (s *DriverService) GetAllDrivers() ([]*models.Driver, error) {
	return s.Repo.GetAllDrivers()
}

func (s *DriverService) GetDriverByID(driverID string) (*models.Driver, error) {
	return s.Repo.FindByID(driverID)
}

func (s *DriverService) UpdateDriver(driver *models.Driver) error {
	return s.Repo.UpdateDriver(driver)
}

func (s *DriverService) GetPendingBookings(driverID string) ([]*models.Booking, error) {
    return s.BookingRepo.FindAssignedBookings(driverID)
}

func (s *DriverService) RespondToBooking(driverID, bookingID, response string) error {
    if response == "accept" {
        return s.BookingService.DriverAcceptsBooking(driverID, bookingID)
    } else if response == "reject" {
        return s.BookingService.DriverRejectsBooking(driverID, bookingID)
    }
    return errors.New("invalid response")
}

// isValidTransition checks if the status transition is valid
func isValidTransition(currentStatus, newStatus string, validTransitions map[string][]string) bool {
	validNextStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, validStatus := range validNextStatuses {
		if newStatus == validStatus {
			return true
		}
	}

	return false
}
