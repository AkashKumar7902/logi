package models

import "time"

type Booking struct {
    ID             string     `bson:"_id,omitempty" json:"id,omitempty"`
    UserID         string     `bson:"user_id" json:"user_id"`
    DriverID       string     `bson:"driver_id,omitempty" json:"driver_id,omitempty"`
    PickupLocation Location   `bson:"pickup_location" json:"pickup_location"`
    DropoffLocation Location  `bson:"dropoff_location" json:"dropoff_location"`
    VehicleType    string     `bson:"vehicle_type" json:"vehicle_type"`
    PriceEstimate  float64    `bson:"price_estimate" json:"price_estimate"`
    Status         string     `bson:"status" json:"status"`
    CreatedAt      time.Time  `bson:"created_at" json:"created_at"`
    ScheduledTime  *time.Time `bson:"scheduled_time,omitempty" json:"scheduled_time,omitempty"`
    StartedAt   *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
    CompletedAt *time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type BookingRequest struct {
    PickupLocation  Location   `json:"pickup_location"`
    DropoffLocation Location   `json:"dropoff_location"`
    VehicleType     string     `json:"vehicle_type"`
    ScheduledTime   *time.Time `json:"scheduled_time,omitempty"`
}

type BookingStatistics struct {
    TotalBookings     int64   `json:"total_bookings"`
    CompletedBookings int64   `json:"completed_bookings"`
    AverageTripTime   float64 `json:"average_trip_time"` // in minutes
}

type PriceEstimateRequest struct {
    PickupLocation  Location `json:"pickup_location" binding:"required"`
    DropoffLocation Location `json:"dropoff_location" binding:"required"`
    VehicleType     string   `json:"vehicle_type" binding:"required"`
}

type PriceEstimateResponse struct {
    EstimatedPrice float64 `json:"estimated_price"`
}

