package services

import (
	"context"
	"logi/internal/models"
	"testing"
)

func TestDriverServiceUpdateBookingStatusCompletedClearsDriverAndPublishes(t *testing.T) {
	t.Parallel()

	var updatedBooking *models.Booking
	completedCount := 0
	var clearedDriverID string
	var clearedBookingID string
	var availableDriverID string
	var availableStatus string

	bookingRepo := &fakeBookingRepository{
		findByIDFn: func(ctx context.Context, id string) (*models.Booking, error) {
			return &models.Booking{
				ID:       id,
				UserID:   "user-1",
				DriverID: "driver-1",
				Status:   "Delivered",
			}, nil
		},
		updateFn: func(ctx context.Context, booking *models.Booking) error {
			copy := *booking
			updatedBooking = &copy
			return nil
		},
	}
	driverRepo := &fakeDriverRepository{
		incrementCompletedFn: func(ctx context.Context, driverID string) error {
			completedCount++
			return nil
		},
		updateCurrentBookingIDFn: func(ctx context.Context, driverID, bookingID string) error {
			clearedDriverID = driverID
			clearedBookingID = bookingID
			return nil
		},
		updateStatusFn: func(ctx context.Context, driverID, status string) error {
			availableDriverID = driverID
			availableStatus = status
			return nil
		},
	}
	messaging := &fakeMessagingClient{}

	service := &DriverService{
		Repo:            driverRepo,
		BookingRepo:     bookingRepo,
		UserRepo:        &fakeUserRepository{},
		MessagingClient: messaging,
	}

	if err := service.UpdateBookingStatus(context.Background(), "driver-1", "booking-1", "Completed"); err != nil {
		t.Fatalf("UpdateBookingStatus returned error: %v", err)
	}

	if updatedBooking == nil {
		t.Fatal("booking update was not persisted")
	}
	if updatedBooking.Status != "Completed" || updatedBooking.CompletedAt == nil {
		t.Fatalf("booking was not completed correctly: %+v", updatedBooking)
	}
	if completedCount != 1 {
		t.Fatalf("expected completed bookings increment once, got %d", completedCount)
	}
	if clearedDriverID != "driver-1" || clearedBookingID != "" {
		t.Fatalf("current booking was not cleared: %s %s", clearedDriverID, clearedBookingID)
	}
	if availableDriverID != "driver-1" || availableStatus != "Available" {
		t.Fatalf("driver availability was not restored: %s %s", availableDriverID, availableStatus)
	}
	if len(messaging.published) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(messaging.published))
	}
	if messaging.published[0].userID != "user-1" || messaging.published[0].messageType != "status_update" {
		t.Fatalf("unexpected user message: %+v", messaging.published[0])
	}
	if messaging.published[1].userID != "" || messaging.published[1].messageType != "driver_status_update" {
		t.Fatalf("unexpected admin message: %+v", messaging.published[1])
	}
}

func TestDriverServiceUpdateBookingStatusRejectsInvalidTransition(t *testing.T) {
	t.Parallel()

	updateCalled := false
	service := &DriverService{
		Repo: &fakeDriverRepository{},
		BookingRepo: &fakeBookingRepository{
			findByIDFn: func(ctx context.Context, id string) (*models.Booking, error) {
				return &models.Booking{
					ID:       id,
					UserID:   "user-1",
					DriverID: "driver-1",
					Status:   "Driver Assigned",
				}, nil
			},
			updateFn: func(ctx context.Context, booking *models.Booking) error {
				updateCalled = true
				return nil
			},
		},
		UserRepo:        &fakeUserRepository{},
		MessagingClient: &fakeMessagingClient{},
	}

	err := service.UpdateBookingStatus(context.Background(), "driver-1", "booking-1", "Completed")
	if err == nil {
		t.Fatal("expected invalid transition error")
	}
	if updateCalled {
		t.Fatal("booking should not be updated for an invalid transition")
	}
}
