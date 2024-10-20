package services

import (
	"logi/internal/models"
	"logi/internal/repositories"
	"logi/internal/services/distance"
	"math"
	"time"
)

type PricingService struct {
	BookingRepo  repositories.BookingRepository
	DriverRepo   repositories.DriverRepository
	DistanceCalc distance.DistanceCalculator
}

func NewPricingService(bookingRepo repositories.BookingRepository, driverRepo repositories.DriverRepository, distanceCalc distance.DistanceCalculator) *PricingService {
	return &PricingService{
		BookingRepo:  bookingRepo,
		DriverRepo:   driverRepo,
		DistanceCalc: distanceCalc,
	}
}

// CalculatePrice calculates the price based on pickup and dropoff locations and vehicle type.
func (s *PricingService) CalculatePrice(pickup, dropoff models.Location, vehicleType string) (float64, error) {
	basePrice, err := s.calculateBasePrice(pickup, dropoff, vehicleType)
	if err != nil {
		return 0, err
	}

	surgeMultiplier := s.calculateSurgeMultiplier()

	finalPrice := basePrice * surgeMultiplier
	return math.Round(finalPrice*100) / 100, nil // Round to two decimal places
}

func (s *PricingService) calculateBasePrice(pickup, dropoff models.Location, vehicleType string) (float64, error) {
	distanceResult, err := s.DistanceCalc.Calculate(pickup, dropoff)
	if err != nil {
		return 0, err
	}

	ratePerKm := s.getRatePerKm(vehicleType)
	return distanceResult.Distance * ratePerKm, nil
}

func (s *PricingService) calculateDistance(pickup, dropoff models.Location) float64 {
	pickupLat := pickup.Coordinates[1]
	pickupLon := pickup.Coordinates[0]
	dropoffLat := dropoff.Coordinates[1]
	dropoffLon := dropoff.Coordinates[0]

	// Haversine formula to calculate distance between two points
	const EarthRadius = 6371 // Kilometers

	dLat := degreesToRadians(dropoffLat - pickupLat)
	dLon := degreesToRadians(dropoffLon - pickupLon)

	lat1 := degreesToRadians(pickupLat)
	lat2 := degreesToRadians(dropoffLat)

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
		return 6.0
	case "car":
		return 12.0
	case "van":
		return 18.0
	default:
		return 30.0
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
