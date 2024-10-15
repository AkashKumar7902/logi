package services

import (
    "errors"
    "logi/internal/models"
    "logi/internal/repositories"
    "time"

    "github.com/google/uuid"
)

type BookingService struct {
    Repo           repositories.BookingRepository
    DriverRepo     repositories.DriverRepository
    PricingService *PricingService
}

func NewBookingService(repo repositories.BookingRepository, driverRepo repositories.DriverRepository, pricingService *PricingService) *BookingService {
    return &BookingService{
        Repo:           repo,
        DriverRepo:     driverRepo,
        PricingService: pricingService,
    }
}

func (s *BookingService) CreateBooking(userID string, bookingReq *models.BookingRequest) (*models.Booking, error) {
    // Calculate price with surge pricing
    price := s.PricingService.CalculatePrice(bookingReq.PickupLocation, bookingReq.DropoffLocation, bookingReq.VehicleType)

    booking := &models.Booking{
        ID:             uuid.NewString(),
        UserID:         userID,
        PickupLocation: bookingReq.PickupLocation,
        DropoffLocation: bookingReq.DropoffLocation,
        VehicleType:    bookingReq.VehicleType,
        PriceEstimate:  price,
        Status:         "Pending",
        CreatedAt:      time.Now(),
    }

    if bookingReq.ScheduledTime != nil {
        booking.ScheduledTime = bookingReq.ScheduledTime
    } else {
        // Assign driver immediately
        driver, err := s.DriverRepo.FindAvailableDriver(bookingReq.PickupLocation, bookingReq.VehicleType)
        if err != nil || driver == nil {
            return nil, errors.New("no available drivers")
        }
        booking.DriverID = driver.ID
        booking.Status = "Driver Assigned"

        // Update driver status
        s.DriverRepo.UpdateStatus(driver.ID, "Busy")
    }

    err := s.Repo.Create(booking)
    if err != nil {
        return nil, err
    }

    return booking, nil
}

func (s *BookingService) ActivateScheduledBookings() error {
    bookings, err := s.Repo.FindPendingScheduledBookings()
    if err != nil {
        return err
    }

    for _, booking := range bookings {
        driver, err := s.DriverRepo.FindAvailableDriver(booking.PickupLocation, booking.VehicleType)
        if err != nil || driver == nil {
            continue // Skip if no drivers are available
        }

        booking.DriverID = driver.ID
        booking.Status = "Driver Assigned"

        // Update booking and driver status
        s.Repo.Update(booking)
        s.DriverRepo.UpdateStatus(driver.ID, "Busy")
    }

    return nil
}
