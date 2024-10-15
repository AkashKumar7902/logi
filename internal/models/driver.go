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
}

type Location struct {
    Latitude  float64 `bson:"latitude" json:"latitude"`
    Longitude float64 `bson:"longitude" json:"longitude"`
}
