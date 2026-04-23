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

type AdminService struct {
	Repo        repositories.AdminRepository
	AuthService *auth.AuthService
	UserRepo    repositories.UserRepository
	DriverRepo  repositories.DriverRepository
	BookingRepo repositories.BookingRepository
}

func NewAdminService(repo repositories.AdminRepository, authService *auth.AuthService, userRepo repositories.UserRepository, driverRepo repositories.DriverRepository, bookingRepo repositories.BookingRepository) *AdminService {
	return &AdminService{
		Repo:        repo,
		AuthService: authService,
		UserRepo:    userRepo,
		DriverRepo:  driverRepo,
		BookingRepo: bookingRepo,
	}
}

func (s *AdminService) Register(ctx context.Context, admin *models.Admin, password string) error {
	existingAdmin, _ := s.Repo.FindByEmail(ctx, admin.Email)
	if existingAdmin != nil {
		return errors.New("admin already exists")
	}

	hashedPassword, err := s.AuthService.HashPassword(password)
	if err != nil {
		return err
	}

	admin.ID = uuid.NewString()
	admin.PasswordHash = hashedPassword
	admin.CreatedAt = time.Now()

	return s.Repo.Create(ctx, admin)
}

func (s *AdminService) Login(ctx context.Context, email, password string) (*models.Admin, error) {
	admin, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !s.AuthService.CheckPasswordHash(password, admin.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return admin, nil
}

func (s *AdminService) GetStatistics(ctx context.Context) (*models.AdminStatistics, error) {
	avgTripTime, err := s.BookingRepo.GetAverageTripTime(ctx)
	if err != nil {
		return nil, err
	}

	totalBookings, err := s.BookingRepo.GetTotalBookings(ctx)
	if err != nil {
		return nil, err
	}

	totalDrivers, err := s.DriverRepo.GetTotalDrivers(ctx)
	if err != nil {
		return nil, err
	}

	totalUsers, err := s.UserRepo.GetTotalUsers(ctx)
	if err != nil {
		return nil, err
	}

	stats := &models.AdminStatistics{
		AverageTripTime: avgTripTime,
		TotalBookings:   totalBookings,
		TotalDrivers:    totalDrivers,
		TotalUsers:      totalUsers,
	}

	return stats, nil
}
