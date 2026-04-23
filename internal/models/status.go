package models

const (
	BookingStatusPending         = "Pending"
	BookingStatusDriverAssigned  = "Driver Assigned"
	BookingStatusEnRouteToPickup = "En Route to Pickup"
	BookingStatusGoodsCollected  = "Goods Collected"
	BookingStatusInTransit       = "In Transit"
	BookingStatusDelivered       = "Delivered"
	BookingStatusCompleted       = "Completed"
)

const (
	DriverStatusAvailable = "Available"
	DriverStatusBusy      = "Busy"
	DriverStatusOffline   = "Offline"
)

var ActiveDemandBookingStatuses = []string{
	BookingStatusPending,
	BookingStatusDriverAssigned,
	BookingStatusEnRouteToPickup,
	BookingStatusGoodsCollected,
	BookingStatusInTransit,
	BookingStatusDelivered,
}

var DriverInProgressBookingStatuses = []string{
	BookingStatusDriverAssigned,
	BookingStatusEnRouteToPickup,
	BookingStatusGoodsCollected,
	BookingStatusInTransit,
}
