package scheduler

import (
    "log"
    "logi/internal/services"

    "github.com/robfig/cron/v3"
)

func StartScheduler(bookingService *services.BookingService) {
    c := cron.New()
    c.AddFunc("@every 1m", func() {
        log.Println("Checking for scheduled bookings...")
        err := bookingService.ActivateScheduledBookings()
        if err != nil {
            log.Println("Error activating scheduled bookings:", err)
        }
    })
    c.Start()
}
