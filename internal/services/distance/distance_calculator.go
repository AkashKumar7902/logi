package distance

import "logi/internal/models"

// DistanceResult holds the distance and duration between two locations.
type DistanceResult struct {
    Distance float64 // in kilometers
    Duration float64 // in minutes
}

// DistanceCalculator defines the interface for distance and duration calculations.
type DistanceCalculator interface {
    Calculate(pickup, dropoff models.Location) (*DistanceResult, error)
}
