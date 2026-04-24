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
	if err := models.ValidateBookingRequest(bookingReq); err != nil {
		return nil, err
	}

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
		Status:               models.BookingStatusPending,
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
	excluded := make(map[string]struct{}, len(booking.RejectedDriverIDs)+len(excludedDriverIDs))
	for _, driverID := range booking.RejectedDriverIDs {
		excluded[driverID] = struct{}{}
	}
	for driverID := range excludedDriverIDs {
		excluded[driverID] = struct{}{}
	}

	drivers, err := s.DriverRepo.FindAvailableDrivers(
		ctx,
		booking.PickupLocation,
		booking.VehicleType,
	)

	if err != nil || len(drivers) == 0 {
		utils.Warn(ctx, "no available drivers", "booking_id", booking.ID, "vehicle_type", booking.VehicleType, "error", err)
		return errors.New("no available drivers")
	}

	recipientDrivers := make([]*models.Driver, 0, len(drivers))
	for _, driver := range drivers {
		if _, skip := excluded[driver.ID]; skip {
			continue
		}
		recipientDrivers = append(recipientDrivers, driver)
	}

	booking.OfferedDriverIDs = make([]string, 0, len(recipientDrivers))
	for _, driver := range recipientDrivers {
		booking.OfferedDriverIDs = append(booking.OfferedDriverIDs, driver.ID)
	}
	if err := s.Repo.Update(ctx, booking); err != nil {
		return err
	}

	if len(recipientDrivers) == 0 {
		utils.Warn(ctx, "booking request had no eligible recipients", "booking_id", booking.ID)
		return errors.New("no eligible drivers available")
	}

	publishedCount := 0
	for _, driver := range recipientDrivers {
		err := s.MessagingClient.Publish(driver.ID, "new_booking_request", booking)
		if err != nil {
			utils.Warn(ctx, "failed to send booking request to driver", "booking_id", booking.ID, "driver_id", driver.ID, "error", err)
			continue
		}
		publishedCount++
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
		booking.DriverResponseStatus = "Pending"

		err = s.AssignBookingToDrivers(ctx, booking)
		if err != nil {
			utils.Error(ctx, "failed to assign scheduled booking to drivers", "booking_id", booking.ID, "error", err)
			continue
		}

		now := time.Now()
		booking.ScheduledActivatedAt = &now
		if err := s.Repo.Update(ctx, booking); err != nil {
			utils.Error(ctx, "failed to mark scheduled booking activated", "booking_id", booking.ID, "error", err)
		}
	}

	return nil
}

func (s *BookingService) GetPriceEstimate(ctx context.Context, bookingReq *models.PriceEstimateRequest) (float64, error) {
	if err := models.ValidatePriceEstimateRequest(bookingReq); err != nil {
		return 0, err
	}

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
	booking, err := s.Repo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}

	reserved, err := s.DriverRepo.TryAssignCurrentBooking(ctx, driverID, bookingID)
	if err != nil {
		return err
	}
	if !reserved {
		return errors.New("driver is not available")
	}

	assigned, err := s.Repo.AssignDriverIfUnassigned(ctx, bookingID, driverID)
	if err != nil {
		s.rollbackDriverReservation(ctx, driverID, bookingID)
		return err
	}
	if !assigned {
		s.rollbackDriverReservation(ctx, driverID, bookingID)
		return errors.New("booking already accepted or unavailable")
	}

	// Increment driver's counts
	err = s.DriverRepo.IncrementAcceptedBookings(ctx, driverID)
	if err != nil {
		// Log the error but do not fail the operation
		utils.Warn(ctx, "failed to increment accepted bookings", "driver_id", driverID, "booking_id", bookingID, "error", err)
	}

	// Publish the status update to admins via MessagingClient
	publishErr := s.MessagingClient.Publish("", "driver_status_update", map[string]interface{}{
		"driver_id": driverID,
		"status":    models.DriverStatusBusy,
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

	if booking.DriverID != "" || booking.Status != models.BookingStatusPending {
		return errors.New("booking already accepted or unavailable")
	}
	if !bookingWasOfferedToDriver(booking, driverID) {
		return errors.New("booking was not offered to this driver")
	}

	alreadyRejected := false
	for _, rejectedDriverID := range booking.RejectedDriverIDs {
		if rejectedDriverID == driverID {
			alreadyRejected = true
			break
		}
	}
	if !alreadyRejected {
		booking.RejectedDriverIDs = append(booking.RejectedDriverIDs, driverID)
	}

	// Reassign booking to other eligible drivers, excluding rejecting driver.
	err = s.assignBookingToDrivers(ctx, booking, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *BookingService) rollbackDriverReservation(ctx context.Context, driverID, bookingID string) {
	if err := s.DriverRepo.ClearCurrentBookingIfMatches(ctx, driverID, bookingID); err != nil {
		utils.Warn(ctx, "failed to rollback driver reservation", "driver_id", driverID, "booking_id", bookingID, "error", err)
	}
}

func bookingWasOfferedToDriver(booking *models.Booking, driverID string) bool {
	for _, offeredDriverID := range booking.OfferedDriverIDs {
		if offeredDriverID == driverID {
			return true
		}
	}
	return false
}
