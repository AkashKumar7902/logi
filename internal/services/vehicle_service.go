package services

import (
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

func (s *VehicleService) CreateVehicle(vehicle *models.Vehicle) error {
	vehicle.ID = uuid.NewString()
	vehicle.CreatedAt = time.Now()
	vehicle.UpdatedAt = time.Now()
	return s.Repo.Create(vehicle)
}

func (s *VehicleService) UpdateVehicle(vehicle *models.Vehicle) error {
	vehicle.UpdatedAt = time.Now()
	return s.Repo.Update(vehicle)
}

func (s *VehicleService) DeleteVehicle(vehicleID string) error {
	return s.Repo.Delete(vehicleID)
}

func (s *VehicleService) GetVehicleByID(vehicleID string) (*models.Vehicle, error) {
	return s.Repo.FindByID(vehicleID)
}

func (s *VehicleService) GetAllVehicles() ([]*models.Vehicle, error) {
	return s.Repo.FindAll()
}
