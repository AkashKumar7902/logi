package distance

import (
	"logi/internal/models"
	"math"
)

// HaversineCalculator implements the DistanceCalculator interface using the Haversine formula.
type HaversineCalculator struct{}

// NewHaversineCalculator returns a new instance of HaversineCalculator.
func NewHaversineCalculator() *HaversineCalculator {
	return &HaversineCalculator{}
}

// Calculate computes the distance and duration using the Haversine formula.
func (h *HaversineCalculator) Calculate(pickup, dropoff models.Location) (*DistanceResult, error) {
	distance := haversineDistance(pickup.Coordinates[1], pickup.Coordinates[0], dropoff.Coordinates[1], dropoff.Coordinates[0])
	// Assume average speed of 40 km/h for duration estimation
	duration := (distance / 40.0) * 60.0
	return &DistanceResult{
		Distance: distance,
		Duration: duration,
	}, nil
}

// haversineDistance calculates the distance between two points in kilometers.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const EarthRadius = 6371 // Kilometers
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadius * c
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
