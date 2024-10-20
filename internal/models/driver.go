package models

import "time"

type Driver struct {
    ID           string    `bson:"_id,omitempty" json:"id,omitempty"`
    Name         string    `bson:"name" json:"name"`
    Email        string    `bson:"email" json:"email"`
    PasswordHash string    `bson:"password_hash" json:"-"`
    VehicleType string    `bson:"vehicle_type" json:"vehicle_type"`
    VehicleID       string    `bson:"vehicle_id,omitempty" json:"vehicle_id,omitempty"`
    Location     Location  `bson:"location" json:"location"`
    Status       string    `bson:"status" json:"status"` // Status: Available, Busy, Offline
    CreatedAt    time.Time `bson:"created_at" json:"created_at"`
    CurrentBookingID string `bson:"current_booking_id,omitempty" json:"current_booking_id,omitempty"`
    AcceptedBookingsCount int `bson:"accepted_bookings_count" json:"accepted_bookings_count"`
    TotalBookingsCount    int `bson:"total_bookings_count" json:"total_bookings_count"`
    CompletedBookingsCount int `bson:"completed_bookings_count" json:"completed_bookings_count"`
}

type Location struct {
    Type        string    `bson:"type" json:"type"`
    Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}
