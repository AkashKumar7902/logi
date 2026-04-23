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

	var currentBookingDriverID string
	var currentBookingID string
	var busyDriverID string
	var busyStatus string
	acceptedCount := 0

	driverRepo := &fakeDriverRepository{
		updateCurrentBookingIDFn: func(ctx context.Context, driverID, bookingID string) error {
			currentBookingDriverID = driverID
			currentBookingID = bookingID
			return nil
		},
		incrementAcceptedFn: func(ctx context.Context, driverID string) error {
			if driverID != "driver-1" {
				t.Fatalf("unexpected driver increment: %s", driverID)
			}
			acceptedCount++
			return nil
		},
		updateStatusFn: func(ctx context.Context, driverID, status string) error {
			busyDriverID = driverID
			busyStatus = status
			return nil
		},
	}

	messaging := &fakeMessagingClient{}
	service := NewBookingService(bookingRepo, driverRepo, nil, messaging)

	if err := service.DriverAcceptsBooking(context.Background(), "driver-1", "booking-1"); err != nil {
		t.Fatalf("DriverAcceptsBooking returned error: %v", err)
	}

	if currentBookingDriverID != "driver-1" || currentBookingID != "booking-1" {
		t.Fatalf("current booking was not assigned correctly: %s %s", currentBookingDriverID, currentBookingID)
	}
	if acceptedCount != 1 {
		t.Fatalf("expected accepted count increment once, got %d", acceptedCount)
	}
	if busyDriverID != "driver-1" || busyStatus != "Busy" {
		t.Fatalf("driver status not updated to busy: %s %s", busyDriverID, busyStatus)
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

func TestBookingServiceDriverAcceptsBookingReturnsErrorWhenUnavailable(t *testing.T) {
	t.Parallel()

	driverRepoCalled := false
	service := NewBookingService(
		&fakeBookingRepository{
			assignDriverIfUnassignedFn: func(ctx context.Context, bookingID, driverID string) (bool, error) {
				return false, nil
			},
		},
		&fakeDriverRepository{
			updateCurrentBookingIDFn: func(ctx context.Context, driverID, bookingID string) error {
				driverRepoCalled = true
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
	if driverRepoCalled {
		t.Fatal("driver repository should not be updated when assignment fails")
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
				VehicleType: "car",
				Status:      "Pending",
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
					VehicleType: "car",
					Status:      "Pending",
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
