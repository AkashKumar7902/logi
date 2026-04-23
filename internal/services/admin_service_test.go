package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"logi/internal/models"
	"logi/pkg/auth"

	"go.mongodb.org/mongo-driver/mongo"
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

func TestAdminServiceRegisterBootstrapCreatesFirstAdmin(t *testing.T) {
	t.Parallel()

	var createdAdmin *models.Admin
	authService := auth.NewAuthService(strings.Repeat("a", 32), 72)
	service := NewAdminService(
		&fakeAdminRepository{
			hasAnyFn: func(ctx context.Context) (bool, error) {
				return false, nil
			},
			findByEmailFn: func(ctx context.Context, email string) (*models.Admin, error) {
				return nil, mongo.ErrNoDocuments
			},
			createFn: func(ctx context.Context, admin *models.Admin) error {
				createdAdmin = admin
				return nil
			},
		},
		authService,
		nil,
		nil,
		nil,
		nil,
	)

	err := service.RegisterBootstrap(context.Background(), &models.Admin{
		Name:  "Admin",
		Email: "admin@example.com",
	}, "super-secret-password")
	if err != nil {
		t.Fatalf("RegisterBootstrap returned error: %v", err)
	}

	if createdAdmin == nil {
		t.Fatal("expected admin to be created")
	}
	if createdAdmin.ID == "" {
		t.Fatal("expected bootstrap admin ID to be assigned")
	}
	if createdAdmin.PasswordHash == "" {
		t.Fatal("expected bootstrap admin password to be hashed")
	}
	if createdAdmin.PasswordHash == "super-secret-password" {
		t.Fatal("expected bootstrap admin password hash to differ from plaintext password")
	}
}

func TestAdminServiceRegisterBootstrapRejectsAfterFirstAdmin(t *testing.T) {
	t.Parallel()

	service := NewAdminService(
		&fakeAdminRepository{
			hasAnyFn: func(ctx context.Context) (bool, error) {
				return true, nil
			},
		},
		auth.NewAuthService(strings.Repeat("a", 32), 72),
		nil,
		nil,
		nil,
		nil,
	)

	err := service.RegisterBootstrap(context.Background(), &models.Admin{
		Name:  "Admin",
		Email: "admin@example.com",
	}, "super-secret-password")
	if !errors.Is(err, ErrAdminBootstrapAlreadyCompleted) {
		t.Fatalf("expected ErrAdminBootstrapAlreadyCompleted, got %v", err)
	}
}
