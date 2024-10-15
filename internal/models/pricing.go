package models

type PricingFactors struct {
    DemandFactor      float64 `json:"demand_factor"`
    SupplyFactor      float64 `json:"supply_factor"`
    TimeOfDayFactor   float64 `json:"time_of_day_factor"`
    SurgeMultiplier   float64 `json:"surge_multiplier"`
}
