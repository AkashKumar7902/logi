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
}

func NewAdminService(repo repositories.AdminRepository, authService *auth.AuthService) *AdminService {
    return &AdminService{
        Repo:        repo,
        AuthService: authService,
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
