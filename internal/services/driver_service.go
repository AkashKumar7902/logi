package services

import (
	"errors"
	"logi/internal/models"
	"logi/internal/repositories"
	"logi/pkg/auth"
	"time"

	"github.com/google/uuid"
)

type DriverService struct {
	Repo        repositories.DriverRepository
	BookingRepo repositories.BookingRepository
	AuthService *auth.AuthService
}

func NewDriverService(repo repositories.DriverRepository, bookingRepo repositories.BookingRepository, authService *auth.AuthService) *DriverService {
	return &DriverService{
		Repo:        repo,
		AuthService: authService,
		BookingRepo: bookingRepo,
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

func (s *DriverService) UpdateStatus(driverID, status string) error {
	return s.Repo.UpdateStatus(driverID, status)
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

	// If booking is marked as completed, update driver's status to Available
	if status == "Completed" {
		err = s.UpdateStatus(booking.DriverID, "Available")
		if err != nil {
			return errors.New("failed to update driver status to Available")
		}
	}

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
