package services

import (
	"errors"
	"logi/internal/messaging"
	"logi/internal/models"
	"logi/internal/repositories"
	"logi/internal/utils"
	"time"

	"github.com/google/uuid"
)

type BookingService struct {
    Repo           repositories.BookingRepository
    DriverRepo     repositories.DriverRepository
    PricingService *PricingService
    MessagingClient messaging.MessagingClient
}

func NewBookingService(repo repositories.BookingRepository, driverRepo repositories.DriverRepository, pricingService *PricingService, messagingClient messaging.MessagingClient) *BookingService {
    return &BookingService{
        Repo:           repo,
        DriverRepo:     driverRepo,
        PricingService: pricingService,
        MessagingClient: messagingClient,
    }
}

func (s *BookingService) CreateBooking(userID string, bookingReq *models.BookingRequest) (*models.Booking, error) {
    // Calculate price with surge pricing
    price, err := s.PricingService.CalculatePrice(bookingReq.PickupLocation, bookingReq.DropoffLocation, bookingReq.VehicleType)
    if err != nil {
        utils.Logger.Println("failed to calculate price")
        return nil, errors.New("failed to calculate price")
    }

    booking := &models.Booking{
        ID:                uuid.NewString(),
        UserID:            userID,
        PickupLocation:    bookingReq.PickupLocation,
        DropoffLocation:   bookingReq.DropoffLocation,
        VehicleType:       bookingReq.VehicleType,
        PriceEstimate:     price,
        Status:            "Pending",
        DriverResponseStatus: "Pending",
        CreatedAt:         time.Now(),
    }

    if bookingReq.ScheduledTime != nil {
        booking.ScheduledTime = bookingReq.ScheduledTime
    }

    // Save booking to the database
    err = s.Repo.Create(booking)
    if err != nil {
        return nil, err
    }

    if booking.ScheduledTime != nil {
        return booking, nil
    }

    // Assign booking to drivers
    err = s.AssignBookingToDrivers(booking)
    if err != nil {
        return nil, err
    }

    return booking, nil
}

// AssignBookingToDrivers sends booking requests to nearby drivers
func (s *BookingService) AssignBookingToDrivers(booking *models.Booking) error {
    drivers, err := s.DriverRepo.FindAvailableDrivers(
        booking.PickupLocation,
        booking.VehicleType,
    )

    if err != nil || len(drivers) == 0 {
        utils.Logger.Println("no available drivers")
        return errors.New("no available drivers")
    }

    for _, driver := range drivers {
        // Notify each driver
        utils.Logger.Println("sending booking request to driver", driver.ID)
        err := s.MessagingClient.Publish(driver.ID, "new_booking_request", booking)
        if err != nil {
            // Handle messaging errors
            continue
        }
    }
    return nil
}

func (s *BookingService) ActivateScheduledBookings() error {
    bookings, err := s.Repo.FindPendingScheduledBookings()
    if err != nil {
        return err
    }

    for _, booking := range bookings {
        // Update booking status to indicate it's now active and awaiting driver response
        booking.DriverResponseStatus = "Pending"
        err := s.Repo.Update(booking)
        if err != nil {
            utils.Logger.Printf("Failed to update booking status for booking ID %s: %v", booking.ID, err)
            continue
        }

        // Assign booking to nearby drivers
        err = s.AssignBookingToDrivers(booking)
        if err != nil {
            utils.Logger.Printf("Failed to assign booking ID %s to drivers: %v", booking.ID, err)
            // Optionally, you might want to retry or mark the booking as failed
            continue
        }
    }

    return nil
}


func (s *BookingService) GetPriceEstimate(bookingReq *models.PriceEstimateRequest) (float64, error) {
    price, err := s.PricingService.CalculatePrice(
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
func (s *BookingService) DriverAcceptsBooking(driverID, bookingID string) error {
    booking, err := s.Repo.FindByID(bookingID)
    if err != nil {
        return err
    }

    if booking.DriverResponseStatus != "Pending" {
        return errors.New("booking already accepted or rejected")
    }

    // Update booking with driver ID and status
    booking.DriverID = driverID
    booking.Status = "Driver Assigned"
    booking.DriverResponseStatus = "Accepted"
    err = s.Repo.Update(booking)
    if err != nil {
        return err
    }

    // Update driver's current booking
    err = s.DriverRepo.UpdateCurrentBookingID(driverID, bookingID)
    if err != nil {
        return err
    }

    // Increment driver's counts
    err = s.DriverRepo.IncrementAcceptedBookings(driverID)
    if err != nil {
        // Log the error but do not fail the operation
        utils.Logger.Printf("Failed to increment accepted bookings for driver %s: %v", driverID, err)
    }

    err = s.DriverRepo.UpdateStatus(driverID, "Busy")
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
		utils.Logger.Printf("Failed to publish driver status update: %v", publishErr)
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
func (s *BookingService) DriverRejectsBooking(driverID, bookingID string) error {
    booking, err := s.Repo.FindByID(bookingID)
    if err != nil {
        return err
    }

    if booking.DriverResponseStatus != "Pending" {
        return errors.New("booking already accepted or rejected")
    }

    // Update booking status
    booking.DriverResponseStatus = "Rejected"
    err = s.Repo.Update(booking)
    if err != nil {
        return err
    }

    // Optionally, reassign the booking to other drivers
    err = s.AssignBookingToDrivers(booking)
    if err != nil {
        return err
    }

    return nil
}
