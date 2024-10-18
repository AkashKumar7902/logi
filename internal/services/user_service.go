package services

import (
	"errors"
	"logi/internal/models"
	"logi/internal/repositories"
	"logi/internal/utils"
	"logi/pkg/auth"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	Repo        repositories.UserRepository
	BookingRepo repositories.BookingRepository
	DriverRepo  repositories.DriverRepository
	AuthService *auth.AuthService
}

func NewUserService(repo repositories.UserRepository, bookingRepo repositories.BookingRepository, driverRepo repositories.DriverRepository, authService *auth.AuthService) *UserService {
	return &UserService{
		Repo:        repo,
		BookingRepo: bookingRepo,
        DriverRepo: driverRepo,
		AuthService: authService,
	}
}

func (s *UserService) Register(user *models.User, password string) error {
	existingUser, _ := s.Repo.FindByEmail(user.Email)
	if existingUser != nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := s.AuthService.HashPassword(password)
	if err != nil {
		return err
	}

	user.ID = uuid.NewString()
	user.PasswordHash = hashedPassword
	user.CreatedAt = time.Now()
	user.Role = "user" // Add this line

	return s.Repo.Create(user)
}

func (s *UserService) Login(email, password string) (*models.User, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !s.AuthService.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *UserService) GetActiveBooking(userID string) (*models.Booking, error) {
	// Fetch bookings that are not 'Completed' or 'Pending'
	booking, err := s.BookingRepo.GetActiveBookingByUserID(userID)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *UserService) GetDriverForBooking(userID, bookingID string) (*models.Driver, error) {
	// Fetch the booking
	booking, err := s.BookingRepo.FindByID(bookingID)
	if err != nil {
		return nil, err
	}

	// Check if the booking belongs to the user
	if booking.UserID != userID {
		return nil, errors.New("unauthorized access to booking")
	}

	if booking.DriverID == "" {
		return nil, errors.New("no driver assigned to the booking")
	}

	utils.Logger.Println(booking.DriverID)
	// Fetch the driver assigned to the booking
	driver, err := s.DriverRepo.FindByID(booking.DriverID)
	if err != nil {
		return nil, err
	}
	return driver, nil
}
