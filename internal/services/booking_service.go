package services

import (
	"context"
	"errors"
	"logi/internal/messaging"
	"logi/internal/models"
	"logi/internal/repositories"
	"logi/internal/utils"
	"time"

	"github.com/google/uuid"
)

type BookingService struct {
	Repo            repositories.BookingRepository
	DriverRepo      repositories.DriverRepository
	PricingService  *PricingService
	MessagingClient messaging.MessagingClient
}

func NewBookingService(repo repositories.BookingRepository, driverRepo repositories.DriverRepository, pricingService *PricingService, messagingClient messaging.MessagingClient) *BookingService {
	return &BookingService{
		Repo:            repo,
		DriverRepo:      driverRepo,
		PricingService:  pricingService,
		MessagingClient: messagingClient,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID string, bookingReq *models.BookingRequest) (*models.Booking, error) {
	// Calculate price with surge pricing
	price, err := s.PricingService.CalculatePrice(ctx, bookingReq.PickupLocation, bookingReq.DropoffLocation, bookingReq.VehicleType)
	if err != nil {
		utils.Error(ctx, "failed to calculate price", "user_id", userID, "vehicle_type", bookingReq.VehicleType, "error", err)
		return nil, errors.New("failed to calculate price")
	}

	booking := &models.Booking{
		ID:                   uuid.NewString(),
		UserID:               userID,
		PickupLocation:       bookingReq.PickupLocation,
		DropoffLocation:      bookingReq.DropoffLocation,
		VehicleType:          bookingReq.VehicleType,
		PriceEstimate:        price,
		Status:               "Pending",
		DriverResponseStatus: "Pending",
		CreatedAt:            time.Now(),
	}

	if bookingReq.ScheduledTime != nil {
		booking.ScheduledTime = bookingReq.ScheduledTime
	}

	// Save booking to the database
	err = s.Repo.Create(ctx, booking)
	if err != nil {
		return nil, err
	}

	if booking.ScheduledTime != nil {
		return booking, nil
	}

	// Assign booking to drivers
	err = s.AssignBookingToDrivers(ctx, booking)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

// AssignBookingToDrivers sends booking requests to nearby drivers
func (s *BookingService) AssignBookingToDrivers(ctx context.Context, booking *models.Booking) error {
	return s.assignBookingToDrivers(ctx, booking, nil)
}

func (s *BookingService) assignBookingToDrivers(ctx context.Context, booking *models.Booking, excludedDriverIDs map[string]struct{}) error {
	drivers, err := s.DriverRepo.FindAvailableDrivers(
		ctx,
		booking.PickupLocation,
		booking.VehicleType,
	)

	if err != nil || len(drivers) == 0 {
		utils.Warn(ctx, "no available drivers", "booking_id", booking.ID, "vehicle_type", booking.VehicleType, "error", err)
		return errors.New("no available drivers")
	}

	publishedCount := 0
	for _, driver := range drivers {
		if excludedDriverIDs != nil {
			if _, excluded := excludedDriverIDs[driver.ID]; excluded {
				continue
			}
		}

		err := s.MessagingClient.Publish(driver.ID, "new_booking_request", booking)
		if err != nil {
			utils.Warn(ctx, "failed to send booking request to driver", "booking_id", booking.ID, "driver_id", driver.ID, "error", err)
			continue
		}
		publishedCount++
	}

	if publishedCount == 0 {
		utils.Warn(ctx, "booking request had no eligible recipients", "booking_id", booking.ID)
		return errors.New("no eligible drivers available")
	}

	utils.Info(ctx, "booking request dispatched", "booking_id", booking.ID, "recipient_count", publishedCount)
	return nil
}

func (s *BookingService) ActivateScheduledBookings(ctx context.Context) error {
	bookings, err := s.Repo.FindPendingScheduledBookings(ctx)
	if err != nil {
		return err
	}

	for _, booking := range bookings {
		// Update booking status to indicate it's now active and awaiting driver response
		booking.DriverResponseStatus = "Pending"
		err := s.Repo.Update(ctx, booking)
		if err != nil {
			utils.Error(ctx, "failed to update scheduled booking status", "booking_id", booking.ID, "error", err)
			continue
		}

		// Assign booking to nearby drivers
		err = s.AssignBookingToDrivers(ctx, booking)
		if err != nil {
			utils.Error(ctx, "failed to assign scheduled booking to drivers", "booking_id", booking.ID, "error", err)
			// Optionally, you might want to retry or mark the booking as failed
			continue
		}
	}

	return nil
}

func (s *BookingService) GetPriceEstimate(ctx context.Context, bookingReq *models.PriceEstimateRequest) (float64, error) {
	price, err := s.PricingService.CalculatePrice(
		ctx,
		bookingReq.PickupLocation,
		bookingReq.DropoffLocation,
		bookingReq.VehicleType,
	)
	if err != nil {
		return 0, err
	}
	return price, nil
}

// DriverAcceptsBooking handles driver's acceptance
func (s *BookingService) DriverAcceptsBooking(ctx context.Context, driverID, bookingID string) error {
	assigned, err := s.Repo.AssignDriverIfUnassigned(ctx, bookingID, driverID)
	if err != nil {
		return err
	}
	if !assigned {
		return errors.New("booking already accepted or unavailable")
	}

	booking, err := s.Repo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	// Update driver's current booking
	err = s.DriverRepo.UpdateCurrentBookingID(ctx, driverID, bookingID)
	if err != nil {
		return err
	}

	// Increment driver's counts
	err = s.DriverRepo.IncrementAcceptedBookings(ctx, driverID)
	if err != nil {
		// Log the error but do not fail the operation
		utils.Warn(ctx, "failed to increment accepted bookings", "driver_id", driverID, "booking_id", bookingID, "error", err)
	}

	err = s.DriverRepo.UpdateStatus(ctx, driverID, "Busy")
	if err != nil {
		return err
	}

	// Publish the status update to admins via MessagingClient
	publishErr := s.MessagingClient.Publish("", "driver_status_update", map[string]interface{}{
		"driver_id": driverID,
		"status":    "Busy",
	})
	if publishErr != nil {
		// Log the error but do not fail the operation
		utils.Warn(ctx, "failed to publish driver status update", "driver_id", driverID, "error", publishErr)
	}

	// Notify user that a driver has accepted the booking
	err = s.MessagingClient.Publish(booking.UserID, "booking_accepted", map[string]interface{}{
		"booking_id": booking.ID,
		"driver_id":  driverID,
	})
	if err != nil {
		// Handle messaging errors
	}

	return nil
}

// DriverRejectsBooking handles driver's rejection
func (s *BookingService) DriverRejectsBooking(ctx context.Context, driverID, bookingID string) error {
	booking, err := s.Repo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.DriverID != "" || booking.Status != "Pending" {
		return errors.New("booking already accepted or unavailable")
	}

	// Reassign booking to other eligible drivers, excluding rejecting driver.
	err = s.assignBookingToDrivers(ctx, booking, map[string]struct{}{driverID: {}})
	if err != nil {
		return err
	}

	return nil
}
