package services

import (
	"context"
	"logi/internal/models"
	"logi/internal/repositories"
	"time"

	"github.com/google/uuid"
)

type VehicleService struct {
	Repo repositories.VehicleRepository
}

func NewVehicleService(repo repositories.VehicleRepository) *VehicleService {
	return &VehicleService{
		Repo: repo,
	}
}

func (s *VehicleService) CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error {
	vehicle.ID = uuid.NewString()
	vehicle.CreatedAt = time.Now()
	vehicle.UpdatedAt = time.Now()
	return s.Repo.Create(ctx, vehicle)
}

func (s *VehicleService) UpdateVehicle(ctx context.Context, vehicle *models.Vehicle) error {
	vehicle.UpdatedAt = time.Now()
	return s.Repo.Update(ctx, vehicle)
}

func (s *VehicleService) DeleteVehicle(ctx context.Context, vehicleID string) error {
	return s.Repo.Delete(ctx, vehicleID)
}

func (s *VehicleService) GetVehicleByID(ctx context.Context, vehicleID string) (*models.Vehicle, error) {
	return s.Repo.FindByID(ctx, vehicleID)
}

func (s *VehicleService) GetAllVehicles(ctx context.Context) ([]*models.Vehicle, error) {
	return s.Repo.FindAll(ctx)
}
