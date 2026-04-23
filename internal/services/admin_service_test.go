package services

import (
	"context"
	"testing"

	"logi/internal/models"
)

func TestAdminServiceAssignVehicleToDriverUpdatesBothSides(t *testing.T) {
	t.Parallel()

	var assignedDriverID string
	var assignedVehicleID string
	var assignedVehicleType string
	var vehicleAssignmentDriverID string
	var vehicleAssignmentID string

	service := NewAdminService(
		nil,
		nil,
		nil,
		&fakeDriverRepository{
			findByIDFn: func(ctx context.Context, driverID string) (*models.Driver, error) {
				return &models.Driver{
					ID:          driverID,
					VehicleType: "car",
				}, nil
			},
			assignVehicleFn: func(ctx context.Context, driverID, vehicleID, vehicleType string) error {
				assignedDriverID = driverID
				assignedVehicleID = vehicleID
				assignedVehicleType = vehicleType
				return nil
			},
		},
		nil,
		&fakeVehicleRepository{
			findByIDFn: func(ctx context.Context, vehicleID string) (*models.Vehicle, error) {
				return &models.Vehicle{
					ID:          vehicleID,
					VehicleType: "van",
				}, nil
			},
			assignDriverFn: func(ctx context.Context, vehicleID, driverID string) error {
				vehicleAssignmentID = vehicleID
				vehicleAssignmentDriverID = driverID
				return nil
			},
		},
	)

	if err := service.AssignVehicleToDriver(context.Background(), "driver-1", "vehicle-1"); err != nil {
		t.Fatalf("AssignVehicleToDriver returned error: %v", err)
	}

	if assignedDriverID != "driver-1" || assignedVehicleID != "vehicle-1" || assignedVehicleType != "van" {
		t.Fatalf("driver assignment not updated correctly: %s %s %s", assignedDriverID, assignedVehicleID, assignedVehicleType)
	}
	if vehicleAssignmentID != "vehicle-1" || vehicleAssignmentDriverID != "driver-1" {
		t.Fatalf("vehicle assignment not updated correctly: %s %s", vehicleAssignmentID, vehicleAssignmentDriverID)
	}
}
