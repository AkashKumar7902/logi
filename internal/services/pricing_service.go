package services

import (
    "logi/internal/models"
    "logi/internal/repositories"
    "math"
    "time"
)

type PricingService struct {
    BookingRepo repositories.BookingRepository
    DriverRepo  repositories.DriverRepository
}

func NewPricingService(bookingRepo repositories.BookingRepository, driverRepo repositories.DriverRepository) *PricingService {
    return &PricingService{
        BookingRepo: bookingRepo,
        DriverRepo:  driverRepo,
    }
}

func (s *PricingService) CalculatePrice(pickup, dropoff models.Location, vehicleType string) float64 {
    basePrice := s.calculateBasePrice(pickup, dropoff, vehicleType)
    surgeMultiplier := s.calculateSurgeMultiplier()

    finalPrice := basePrice * surgeMultiplier
    return math.Round(finalPrice*100) / 100 // Round to two decimal places
}

func (s *PricingService) calculateBasePrice(pickup, dropoff models.Location, vehicleType string) float64 {
    distance := s.calculateDistance(pickup, dropoff)
    ratePerKm := s.getRatePerKm(vehicleType)
    return distance * ratePerKm
}

func (s *PricingService) calculateDistance(pickup, dropoff models.Location) float64 {
    // Haversine formula to calculate distance between two points
    const EarthRadius = 6371 // Kilometers

    dLat := degreesToRadians(dropoff.Latitude - pickup.Latitude)
    dLon := degreesToRadians(dropoff.Longitude - pickup.Longitude)

    lat1 := degreesToRadians(pickup.Latitude)
    lat2 := degreesToRadians(dropoff.Latitude)

    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

    distance := EarthRadius * c
    return distance
}

func degreesToRadians(degrees float64) float64 {
    return degrees * math.Pi / 180
}

func (s *PricingService) getRatePerKm(vehicleType string) float64 {
    // Example rates
    switch vehicleType {
    case "bike":
        return 1.0
    case "car":
        return 2.0
    case "van":
        return 3.0
    default:
        return 2.0
    }
}

func (s *PricingService) calculateSurgeMultiplier() float64 {
    activeBookings, _ := s.BookingRepo.GetActiveBookingsCount()
    availableDrivers, _ := s.DriverRepo.GetAvailableDriversCount()

    if availableDrivers == 0 {
        return 2.0 // Max surge
    }

    ratio := float64(activeBookings) / float64(availableDrivers)
    if ratio > 1.5 {
        return 1.5
    } else if ratio > 1.0 {
        return 1.2
    }

    // Time-based surge
    hour := time.Now().Hour()
    if hour >= 18 && hour <= 21 { // Peak hours
        return 1.3
    }

    return 1.0
}
