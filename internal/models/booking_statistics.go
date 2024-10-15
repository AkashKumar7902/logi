package models

type BookingStatistics struct {
    TotalBookings     int64   `json:"total_bookings"`
    CompletedBookings int64   `json:"completed_bookings"`
    AverageTripTime   float64 `json:"average_trip_time"` // in minutes
}
