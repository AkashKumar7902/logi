package scheduler

import (
	"context"
	"logi/internal/services"
	"logi/internal/utils"

	"github.com/robfig/cron/v3"
)

func StartScheduler(bookingService *services.BookingService) *cron.Cron {
	c := cron.New()
	_, err := c.AddFunc("@every 1m", func() {
		jobCtx := context.Background()
		utils.Info(jobCtx, "checking scheduled bookings", "component", "scheduler")
		if runErr := bookingService.ActivateScheduledBookings(jobCtx); runErr != nil {
			utils.Error(jobCtx, "failed to activate scheduled bookings", "component", "scheduler", "error", runErr)
		}
	})
	if err != nil {
		utils.ErrorBackground("failed to register scheduler job", "component", "scheduler", "error", err)
	}
	c.Start()
	return c
}
