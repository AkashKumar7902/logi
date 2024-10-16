package models

import "time"

type Vehicle struct {
    ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
    Make        string    `bson:"make" json:"make"`
    Model       string    `bson:"model" json:"model"`
    Year        int       `bson:"year" json:"year"`
    LicensePlate string   `bson:"license_plate" json:"license_plate"`
    VehicleType string    `bson:"vehicle_type" json:"vehicle_type"` // e.g., bike, car, van
	DriverID    string    `bson:"driver_id,omitempty" json:"driver_id,omitempty"` // driver assigned to the vehicle by admin (should be updated explicitly by updatedriver handler endpoint)
    CreatedAt   time.Time `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}
