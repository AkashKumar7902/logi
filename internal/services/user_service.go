package services

import (
	"context"
	"errors"
	"logi/internal/models"
	"logi/internal/repositories"
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
		DriverRepo:  driverRepo,
		AuthService: authService,
	}
}

func (s *UserService) Register(ctx context.Context, user *models.User, password string) error {
	existingUser, _ := s.Repo.FindByEmail(ctx, user.Email)
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

	return s.Repo.Create(ctx, user)
}

func (s *UserService) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !s.AuthService.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *UserService) GetActiveBooking(ctx context.Context, userID string) (*models.Booking, error) {
	// Fetch bookings that are not 'Completed' or 'Pending'
	booking, err := s.BookingRepo.GetActiveBookingByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *UserService) GetDriverForBooking(ctx context.Context, userID, bookingID string) (*models.Driver, error) {
	// Fetch the booking
	booking, err := s.BookingRepo.FindByID(ctx, bookingID)
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

	// Fetch the driver assigned to the booking
	driver, err := s.DriverRepo.FindByID(ctx, booking.DriverID)
	if err != nil {
		return nil, err
	}
	return driver, nil
}
