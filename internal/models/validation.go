package models

import (
	"errors"
	"math"
	"strings"
)

var validVehicleTypes = map[string]struct{}{
	"bike": {},
	"car":  {},
	"van":  {},
}

var validDriverStatuses = map[string]struct{}{
	DriverStatusAvailable: {},
	DriverStatusBusy:      {},
	DriverStatusOffline:   {},
}

func NormalizeVehicleType(vehicleType string) string {
	return strings.ToLower(strings.TrimSpace(vehicleType))
}

func IsValidVehicleType(vehicleType string) bool {
	_, ok := validVehicleTypes[NormalizeVehicleType(vehicleType)]
	return ok
}

func ValidateBookingRequest(req *BookingRequest) error {
	if req == nil {
		return errors.New("booking request is required")
	}
	req.VehicleType = NormalizeVehicleType(req.VehicleType)
	if err := ValidateLocation(req.PickupLocation); err != nil {
		return errors.New("invalid pickup_location: " + err.Error())
	}
	if err := ValidateLocation(req.DropoffLocation); err != nil {
		return errors.New("invalid dropoff_location: " + err.Error())
	}
	if !IsValidVehicleType(req.VehicleType) {
		return errors.New("vehicle_type must be one of: bike, car, van")
	}
	return nil
}

func ValidatePriceEstimateRequest(req *PriceEstimateRequest) error {
	if req == nil {
		return errors.New("price estimate request is required")
	}
	req.VehicleType = NormalizeVehicleType(req.VehicleType)
	if err := ValidateLocation(req.PickupLocation); err != nil {
		return errors.New("invalid pickup_location: " + err.Error())
	}
	if err := ValidateLocation(req.DropoffLocation); err != nil {
		return errors.New("invalid dropoff_location: " + err.Error())
	}
	if !IsValidVehicleType(req.VehicleType) {
		return errors.New("vehicle_type must be one of: bike, car, van")
	}
	return nil
}

func ValidateLocation(location Location) error {
	if location.Type != "Point" {
		return errors.New("type must be Point")
	}
	if len(location.Coordinates) != 2 {
		return errors.New("coordinates must contain longitude and latitude")
	}
	longitude := location.Coordinates[0]
	latitude := location.Coordinates[1]
	return ValidateLatitudeLongitude(latitude, longitude)
}

func ValidateLatitudeLongitude(latitude, longitude float64) error {
	if !isFinite(latitude) || !isFinite(longitude) {
		return errors.New("coordinates must be finite numbers")
	}
	if latitude < -90 || latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if longitude < -180 || longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}

func ValidateDriverStatus(status string) error {
	if _, ok := validDriverStatuses[strings.TrimSpace(status)]; !ok {
		return errors.New("status must be one of: Available, Busy, Offline")
	}
	return nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
