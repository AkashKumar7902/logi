package services

import (
    "errors"
    "logi/internal/models"
    "logi/internal/repositories"
    "logi/pkg/auth"
    "time"

    "github.com/google/uuid"
)

type UserService struct {
    Repo        repositories.UserRepository
    AuthService *auth.AuthService
}

func NewUserService(repo repositories.UserRepository, authService *auth.AuthService) *UserService {
    return &UserService{
        Repo:        repo,
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
