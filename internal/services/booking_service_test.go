package services

import (
	"context"
	"logi/internal/models"
	"testing"
)

func TestBookingServiceDriverAcceptsBookingUpdatesStateAndPublishes(t *testing.T) {
	t.Parallel()

	bookingRepo := &fakeBookingRepository{
		assignDriverIfUnassignedFn: func(ctx context.Context, bookingID, driverID string) (bool, error) {
			if bookingID != "booking-1" || driverID != "driver-1" {
				t.Fatalf("unexpected assignment request: %s %s", bookingID, driverID)
			}
			return true, nil
		},
		findByIDFn: func(ctx context.Context, id string) (*models.Booking, error) {
			return &models.Booking{
				ID:     id,
				UserID: "user-1",
			}, nil
		},
	}

	var reservedDriverID string
	var reservedBookingID string
	acceptedCount := 0

	driverRepo := &fakeDriverRepository{
		incrementAcceptedFn: func(ctx context.Context, driverID string) error {
			if driverID != "driver-1" {
				t.Fatalf("unexpected driver increment: %s", driverID)
			}
			acceptedCount++
			return nil
		},
		tryAssignCurrentBookingFn: func(ctx context.Context, driverID, bookingID string) (bool, error) {
			reservedDriverID = driverID
			reservedBookingID = bookingID
			return true, nil
		},
	}

	messaging := &fakeMessagingClient{}
	service := NewBookingService(bookingRepo, driverRepo, nil, messaging)

	if err := service.DriverAcceptsBooking(context.Background(), "driver-1", "booking-1"); err != nil {
		t.Fatalf("DriverAcceptsBooking returned error: %v", err)
	}

	if reservedDriverID != "driver-1" || reservedBookingID != "booking-1" {
		t.Fatalf("driver reservation was not recorded correctly: %s %s", reservedDriverID, reservedBookingID)
	}
	if acceptedCount != 1 {
		t.Fatalf("expected accepted count increment once, got %d", acceptedCount)
	}
	if len(messaging.published) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(messaging.published))
	}
	if messaging.published[0].messageType != "driver_status_update" || messaging.published[0].userID != "" {
		t.Fatalf("unexpected admin publish: %+v", messaging.published[0])
	}
	if messaging.published[1].messageType != "booking_accepted" || messaging.published[1].userID != "user-1" {
		t.Fatalf("unexpected user publish: %+v", messaging.published[1])
	}
}

func TestBookingServiceDriverAcceptsBookingRollsBackWhenBookingUnavailable(t *testing.T) {
	t.Parallel()

	rollbackCalled := false
	service := NewBookingService(
		&fakeBookingRepository{
			findByIDFn: func(ctx context.Context, id string) (*models.Booking, error) {
				return &models.Booking{
					ID:     id,
					UserID: "user-1",
				}, nil
			},
			assignDriverIfUnassignedFn: func(ctx context.Context, bookingID, driverID string) (bool, error) {
				return false, nil
			},
		},
		&fakeDriverRepository{
			tryAssignCurrentBookingFn: func(ctx context.Context, driverID, bookingID string) (bool, error) {
				return true, nil
			},
			clearCurrentBookingFn: func(ctx context.Context, driverID, bookingID string) error {
				rollbackCalled = true
				if driverID != "driver-1" || bookingID != "booking-1" {
					t.Fatalf("unexpected rollback request: %s %s", driverID, bookingID)
				}
				return nil
			},
		},
		nil,
		&fakeMessagingClient{},
	)

	err := service.DriverAcceptsBooking(context.Background(), "driver-1", "booking-1")
	if err == nil {
		t.Fatal("expected error when booking is unavailable")
	}
	if !rollbackCalled {
		t.Fatal("driver reservation should be rolled back when booking assignment fails")
	}
}

func TestBookingServiceDriverRejectsBookingSkipsRejectingDriver(t *testing.T) {
	t.Parallel()

	bookingRepo := &fakeBookingRepository{
		findByIDFn: func(ctx context.Context, id string) (*models.Booking, error) {
			return &models.Booking{
				ID: id,
				PickupLocation: models.Location{
					Type:        "Point",
					Coordinates: []float64{72.8777, 19.0760},
				},
				VehicleType:      "car",
				Status:           "Pending",
				OfferedDriverIDs: []string{"driver-1", "driver-2"},
			}, nil
		},
	}
	driverRepo := &fakeDriverRepository{
		findAvailableDriversFn: func(ctx context.Context, location models.Location, vehicleType string) ([]*models.Driver, error) {
			return []*models.Driver{
				{ID: "driver-1"},
				{ID: "driver-2"},
			}, nil
		},
	}
	messaging := &fakeMessagingClient{}
	service := NewBookingService(bookingRepo, driverRepo, nil, messaging)

	if err := service.DriverRejectsBooking(context.Background(), "driver-1", "booking-1"); err != nil {
		t.Fatalf("DriverRejectsBooking returned error: %v", err)
	}

	if len(messaging.published) != 1 {
		t.Fatalf("expected 1 published message after rejection, got %d", len(messaging.published))
	}
	if messaging.published[0].userID != "driver-2" || messaging.published[0].messageType != "new_booking_request" {
		t.Fatalf("booking was not reassigned to the next eligible driver: %+v", messaging.published[0])
	}
}

func TestBookingServiceDriverRejectsBookingFailsWhenNoEligibleDriversRemain(t *testing.T) {
	t.Parallel()

	service := NewBookingService(
		&fakeBookingRepository{
			findByIDFn: func(ctx context.Context, id string) (*models.Booking, error) {
				return &models.Booking{
					ID: id,
					PickupLocation: models.Location{
						Type:        "Point",
						Coordinates: []float64{0, 0},
					},
					VehicleType:      "car",
					Status:           "Pending",
					OfferedDriverIDs: []string{"driver-1"},
				}, nil
			},
		},
		&fakeDriverRepository{
			findAvailableDriversFn: func(ctx context.Context, location models.Location, vehicleType string) ([]*models.Driver, error) {
				return []*models.Driver{{ID: "driver-1"}}, nil
			},
		},
		nil,
		&fakeMessagingClient{},
	)

	err := service.DriverRejectsBooking(context.Background(), "driver-1", "booking-1")
	if err == nil || err.Error() != "no eligible drivers available" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBookingServiceDriverRejectsBookingRequiresOffer(t *testing.T) {
	t.Parallel()

	service := NewBookingService(
		&fakeBookingRepository{
			findByIDFn: func(ctx context.Context, id string) (*models.Booking, error) {
				return &models.Booking{
					ID:               id,
					VehicleType:      "car",
					Status:           "Pending",
					OfferedDriverIDs: []string{"driver-2"},
				}, nil
			},
		},
		&fakeDriverRepository{},
		nil,
		&fakeMessagingClient{},
	)

	err := service.DriverRejectsBooking(context.Background(), "driver-1", "booking-1")
	if err == nil || err.Error() != "booking was not offered to this driver" {
		t.Fatalf("unexpected error: %v", err)
	}
}
