package services

import (
	"context"
	"errors"
	"logi/internal/messaging"
	"logi/internal/models"
	"logi/internal/repositories"
	"logi/internal/utils"
	"logi/pkg/auth"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type DriverService struct {
	Repo            repositories.DriverRepository
	BookingRepo     repositories.BookingRepository
	UserRepo        repositories.UserRepository
	BookingService  BookingService
	AuthService     *auth.AuthService
	MessagingClient messaging.MessagingClient
}

func NewDriverService(repo repositories.DriverRepository, bookingRepo repositories.BookingRepository, userRepo repositories.UserRepository, bookingService BookingService, authService *auth.AuthService, messagingClient messaging.MessagingClient) *DriverService {
	return &DriverService{
		Repo:            repo,
		AuthService:     authService,
		BookingRepo:     bookingRepo,
		UserRepo:        userRepo,
		BookingService:  bookingService,
		MessagingClient: messagingClient,
	}
}

func (s *DriverService) Register(ctx context.Context, driver *models.Driver, password string) error {
	existingDriver, _ := s.Repo.FindByEmail(ctx, driver.Email)
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
	driver.AcceptedBookingsCount = 0
	driver.TotalBookingsCount = 0
	driver.CompletedBookingsCount = 0

	return s.Repo.Create(ctx, driver)
}

func (s *DriverService) Login(ctx context.Context, email, password string) (*models.Driver, error) {
	driver, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !s.AuthService.CheckPasswordHash(password, driver.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return driver, nil
}

// UpdateStatus updates the driver's status and notifies admins
func (s *DriverService) UpdateStatus(ctx context.Context, driverID, status string) error {
	// Update the driver's status in the repository
	err := s.Repo.UpdateStatus(ctx, driverID, status)
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
		utils.Warn(ctx, "failed to publish driver status update", "driver_id", driverID, "status", status, "error", publishErr)
	}

	return nil
}

// UpdateBookingStatus updates the status of a booking
func (s *DriverService) UpdateBookingStatus(ctx context.Context, driverID string, bookingID string, status string) error {
	// Find the booking by ID
	booking, err := s.BookingRepo.FindByID(ctx, bookingID)
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
			s.Repo.IncrementCompletedBookings(ctx, driverID)
			booking.CompletedAt = &currentTime
		}
	}

	booking.Status = status
	err = s.BookingRepo.Update(ctx, booking)
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
		clearCurrentBookingErr := s.Repo.UpdateCurrentBookingID(ctx, driverID, "")
		if clearCurrentBookingErr != nil {
			utils.Warn(ctx, "failed to clear current booking", "driver_id", driverID, "booking_id", bookingID, "error", clearCurrentBookingErr)
		}

		err = s.UpdateStatus(ctx, booking.DriverID, "Available")
		if err != nil {
			return errors.New("failed to update driver status to Available")
		}
	}

	return nil
}

func (s *DriverService) UpdateLocation(ctx context.Context, driverID string, latitude, longitude float64) error {
	location := models.Location{
		Type:        "Point",
		Coordinates: []float64{longitude, latitude},
	}

	err := s.Repo.UpdateLocation(ctx, driverID, location)
	if err != nil {
		return err
	}

	// Find current booking for the driver
	booking, err := s.BookingRepo.FindActiveBookingByDriverID(ctx, driverID)
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

func (s *DriverService) GetAllDrivers(ctx context.Context) ([]*models.Driver, error) {
	return s.Repo.GetAllDrivers(ctx)
}

func (s *DriverService) GetDriverByID(ctx context.Context, driverID string) (*models.Driver, error) {
	return s.Repo.FindByID(ctx, driverID)
}

func (s *DriverService) UpdateDriver(ctx context.Context, driver *models.Driver) error {
	return s.Repo.UpdateDriver(ctx, driver)
}

func (s *DriverService) GetPendingBookings(ctx context.Context, driverID string) ([]*models.Booking, error) {
	return s.BookingRepo.FindAssignedBookings(ctx, driverID)
}

func (s *DriverService) RespondToBooking(ctx context.Context, driverID, bookingID, response string) error {
	err := s.Repo.IncrementTotalBookings(ctx, driverID)
	if err != nil {
		// Log the error but do not fail the operation
		utils.Warn(ctx, "failed to increment total bookings", "driver_id", driverID, "booking_id", bookingID, "error", err)
	}
	if response == "accept" {
		return s.BookingService.DriverAcceptsBooking(ctx, driverID, bookingID)
	} else if response == "reject" {
		return s.BookingService.DriverRejectsBooking(ctx, driverID, bookingID)
	}
	return errors.New("invalid response")
}

func (s *DriverService) GetActiveBookings(ctx context.Context, driverID string) ([]*models.Booking, error) {
	// Fetch bookings assigned to the driver that are not 'Completed' or 'Pending'
	bookings, err := s.BookingRepo.GetActiveBookingsByDriverID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (s *DriverService) GetUserForBooking(ctx context.Context, driverID, bookingID string) (*models.User, error) {
	// Fetch the booking
	booking, err := s.BookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// Check if the booking is assigned to the driver
	if booking.DriverID != driverID {
		return nil, errors.New("unauthorized access to booking")
	}

	// Fetch the user who made the booking
	user, err := s.UserRepo.FindByID(ctx, booking.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *DriverService) GetDriverInfo(ctx context.Context, driverID string) (*models.Driver, error) {
	driver, err := s.Repo.FindByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func (s *DriverService) GetBooking(ctx context.Context, driverID, bookingID string) (*models.Booking, error) {
	booking, err := s.BookingRepo.FindByIDAndDriverID(ctx, bookingID, driverID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("booking not found or not assigned to the driver")
		}
		return nil, err
	}
	return booking, nil
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
