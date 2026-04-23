package scheduler

import (
	"context"
	"log"
	"logi/internal/services"

	"github.com/robfig/cron/v3"
)

func StartScheduler(bookingService *services.BookingService) *cron.Cron {
	c := cron.New()
	_, err := c.AddFunc("@every 1m", func() {
		log.Println("Checking for scheduled bookings...")
		if runErr := bookingService.ActivateScheduledBookings(context.Background()); runErr != nil {
			log.Println("Error activating scheduled bookings:", runErr)
		}
	})
	if err != nil {
		log.Println("Failed to register scheduler job:", err)
	}
	c.Start()
	return c
}
