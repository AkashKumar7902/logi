package models

type PriceEstimateRequest struct {
    PickupLocation  Location `json:"pickup_location" binding:"required"`
    DropoffLocation Location `json:"dropoff_location" binding:"required"`
    VehicleType     string   `json:"vehicle_type" binding:"required"`
}

type PriceEstimateResponse struct {
    EstimatedPrice float64 `json:"estimated_price"`
}
