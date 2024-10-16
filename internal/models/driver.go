package models

import "time"

type Driver struct {
    ID           string    `bson:"_id,omitempty" json:"id,omitempty"`
    Name         string    `bson:"name" json:"name"`
    Email        string    `bson:"email" json:"email"`
    PasswordHash string    `bson:"password_hash" json:"-"`
    VehicleType  string    `bson:"vehicle_type" json:"vehicle_type"`
    Location     Location  `bson:"location" json:"location"`
    Status       string    `bson:"status" json:"status"` // Status: Available, Busy, Offline
    CreatedAt    time.Time `bson:"created_at" json:"created_at"`
    CurrentBookingID string `bson:"current_booking_id,omitempty" json:"current_booking_id,omitempty"`
}

type Location struct {
    Type        string    `bson:"type" json:"type"`
    Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}
