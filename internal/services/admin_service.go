package services

import (
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

func (s *AdminService) Register(admin *models.Admin, password string) error {
    existingAdmin, _ := s.Repo.FindByEmail(admin.Email)
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

    return s.Repo.Create(admin)
}

func (s *AdminService) Login(email, password string) (*models.Admin, error) {
    admin, err := s.Repo.FindByEmail(email)
    if err != nil {
        return nil, errors.New("invalid email or password")
    }

    if !s.AuthService.CheckPasswordHash(password, admin.PasswordHash) {
        return nil, errors.New("invalid email or password")
    }

    return admin, nil
}

func (s *AdminService) GetStatistics() (*models.AdminStatistics, error) {
    avgTripTime, err := s.BookingRepo.GetAverageTripTime()
    if err != nil {
        return nil, err
    }

    totalBookings, err := s.BookingRepo.GetTotalBookings()
    if err != nil {
        return nil, err
    }

    totalDrivers, err := s.DriverRepo.GetTotalDrivers()
    if err != nil {
        return nil, err
    }

    totalUsers, err := s.UserRepo.GetTotalUsers()
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
