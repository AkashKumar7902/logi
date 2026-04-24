package services

import (
	"context"
	"logi/internal/models"
)

type publishedMessage struct {
	userID      string
	messageType string
	payload     interface{}
}

type fakeAdminRepository struct {
	createFn      func(context.Context, *models.Admin) error
	findByEmailFn func(context.Context, string) (*models.Admin, error)
	hasAnyFn      func(context.Context) (bool, error)
}

func (f *fakeAdminRepository) Create(ctx context.Context, admin *models.Admin) error {
	if f.createFn != nil {
		return f.createFn(ctx, admin)
	}
	return nil
}

func (f *fakeAdminRepository) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	if f.findByEmailFn != nil {
		return f.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (f *fakeAdminRepository) HasAny(ctx context.Context) (bool, error) {
	if f.hasAnyFn != nil {
		return f.hasAnyFn(ctx)
	}
	return false, nil
}

type fakeMessagingClient struct {
	published  []publishedMessage
	publishErr error
}

func (f *fakeMessagingClient) Publish(userID string, messageType string, payload interface{}) error {
	if f.publishErr != nil {
		return f.publishErr
	}
	f.published = append(f.published, publishedMessage{
		userID:      userID,
		messageType: messageType,
		payload:     payload,
	})
	return nil
}

type fakeBookingRepository struct {
	createFn                   func(context.Context, *models.Booking) error
	updateFn                   func(context.Context, *models.Booking) error
	findByIDFn                 func(context.Context, string) (*models.Booking, error)
	assignDriverIfUnassignedFn func(context.Context, string, string) (bool, error)
	findActiveByDriverIDFn     func(context.Context, string) (*models.Booking, error)
	findPendingScheduledFn     func(context.Context) ([]*models.Booking, error)
	getActiveBookingsCountFn   func(context.Context) (int64, error)
	findAssignedBookingsFn     func(context.Context, string) ([]*models.Booking, error)
	updateDriverResponseFn     func(context.Context, string, string) error
	getActiveByDriverIDFn      func(context.Context, string) ([]*models.Booking, error)
	getActiveByUserIDFn        func(context.Context, string) (*models.Booking, error)
	findByIDAndDriverIDFn      func(context.Context, string, string) (*models.Booking, error)
	getAverageTripTimeFn       func(context.Context) (float64, error)
	getTotalBookingsFn         func(context.Context) (int64, error)
}

func (f *fakeBookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	if f.createFn != nil {
		return f.createFn(ctx, booking)
	}
	return nil
}

func (f *fakeBookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, booking)
	}
	return nil
}

func (f *fakeBookingRepository) FindByID(ctx context.Context, id string) (*models.Booking, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeBookingRepository) AssignDriverIfUnassigned(ctx context.Context, bookingID, driverID string) (bool, error) {
	if f.assignDriverIfUnassignedFn != nil {
		return f.assignDriverIfUnassignedFn(ctx, bookingID, driverID)
	}
	return false, nil
}

func (f *fakeBookingRepository) FindActiveBookingByDriverID(ctx context.Context, driverID string) (*models.Booking, error) {
	if f.findActiveByDriverIDFn != nil {
		return f.findActiveByDriverIDFn(ctx, driverID)
	}
	return nil, nil
}

func (f *fakeBookingRepository) FindPendingScheduledBookings(ctx context.Context) ([]*models.Booking, error) {
	if f.findPendingScheduledFn != nil {
		return f.findPendingScheduledFn(ctx)
	}
	return nil, nil
}

func (f *fakeBookingRepository) GetActiveBookingsCount(ctx context.Context) (int64, error) {
	if f.getActiveBookingsCountFn != nil {
		return f.getActiveBookingsCountFn(ctx)
	}
	return 0, nil
}

func (f *fakeBookingRepository) FindAssignedBookings(ctx context.Context, driverID string) ([]*models.Booking, error) {
	if f.findAssignedBookingsFn != nil {
		return f.findAssignedBookingsFn(ctx, driverID)
	}
	return nil, nil
}

func (f *fakeBookingRepository) UpdateDriverResponseStatus(ctx context.Context, bookingID, status string) error {
	if f.updateDriverResponseFn != nil {
		return f.updateDriverResponseFn(ctx, bookingID, status)
	}
	return nil
}

func (f *fakeBookingRepository) GetActiveBookingsByDriverID(ctx context.Context, driverID string) ([]*models.Booking, error) {
	if f.getActiveByDriverIDFn != nil {
		return f.getActiveByDriverIDFn(ctx, driverID)
	}
	return nil, nil
}

func (f *fakeBookingRepository) GetActiveBookingByUserID(ctx context.Context, userID string) (*models.Booking, error) {
	if f.getActiveByUserIDFn != nil {
		return f.getActiveByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeBookingRepository) FindByIDAndDriverID(ctx context.Context, id string, driverID string) (*models.Booking, error) {
	if f.findByIDAndDriverIDFn != nil {
		return f.findByIDAndDriverIDFn(ctx, id, driverID)
	}
	return nil, nil
}

func (f *fakeBookingRepository) GetAverageTripTime(ctx context.Context) (float64, error) {
	if f.getAverageTripTimeFn != nil {
		return f.getAverageTripTimeFn(ctx)
	}
	return 0, nil
}

func (f *fakeBookingRepository) GetTotalBookings(ctx context.Context) (int64, error) {
	if f.getTotalBookingsFn != nil {
		return f.getTotalBookingsFn(ctx)
	}
	return 0, nil
}

type fakeDriverRepository struct {
	createFn                   func(context.Context, *models.Driver) error
	findByEmailFn              func(context.Context, string) (*models.Driver, error)
	findAvailableDriversFn     func(context.Context, models.Location, string) ([]*models.Driver, error)
	updateStatusFn             func(context.Context, string, string) error
	assignVehicleFn            func(context.Context, string, string, string) error
	getAvailableDriversCountFn func(context.Context) (int64, error)
	getAllDriversFn            func(context.Context) ([]*models.Driver, error)
	findByIDFn                 func(context.Context, string) (*models.Driver, error)
	updateDriverFn             func(context.Context, *models.Driver) error
	updateLocationFn           func(context.Context, string, models.Location) error
	updateCurrentBookingIDFn   func(context.Context, string, string) error
	tryAssignCurrentBookingFn  func(context.Context, string, string) (bool, error)
	clearCurrentBookingFn      func(context.Context, string, string) error
	incrementAcceptedFn        func(context.Context, string) error
	incrementTotalFn           func(context.Context, string) error
	incrementCompletedFn       func(context.Context, string) error
	getTotalDriversFn          func(context.Context) (int64, error)
}

func (f *fakeDriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	if f.createFn != nil {
		return f.createFn(ctx, driver)
	}
	return nil
}

func (f *fakeDriverRepository) FindByEmail(ctx context.Context, email string) (*models.Driver, error) {
	if f.findByEmailFn != nil {
		return f.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (f *fakeDriverRepository) FindAvailableDrivers(ctx context.Context, location models.Location, vehicleType string) ([]*models.Driver, error) {
	if f.findAvailableDriversFn != nil {
		return f.findAvailableDriversFn(ctx, location, vehicleType)
	}
	return nil, nil
}

func (f *fakeDriverRepository) UpdateStatus(ctx context.Context, driverID string, status string) error {
	if f.updateStatusFn != nil {
		return f.updateStatusFn(ctx, driverID, status)
	}
	return nil
}

func (f *fakeDriverRepository) AssignVehicle(ctx context.Context, driverID, vehicleID, vehicleType string) error {
	if f.assignVehicleFn != nil {
		return f.assignVehicleFn(ctx, driverID, vehicleID, vehicleType)
	}
	return nil
}

func (f *fakeDriverRepository) GetAvailableDriversCount(ctx context.Context) (int64, error) {
	if f.getAvailableDriversCountFn != nil {
		return f.getAvailableDriversCountFn(ctx)
	}
	return 0, nil
}

func (f *fakeDriverRepository) GetAllDrivers(ctx context.Context) ([]*models.Driver, error) {
	if f.getAllDriversFn != nil {
		return f.getAllDriversFn(ctx)
	}
	return nil, nil
}

func (f *fakeDriverRepository) FindByID(ctx context.Context, driverID string) (*models.Driver, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, driverID)
	}
	return nil, nil
}

func (f *fakeDriverRepository) UpdateDriver(ctx context.Context, driver *models.Driver) error {
	if f.updateDriverFn != nil {
		return f.updateDriverFn(ctx, driver)
	}
	return nil
}

func (f *fakeDriverRepository) UpdateLocation(ctx context.Context, driverID string, location models.Location) error {
	if f.updateLocationFn != nil {
		return f.updateLocationFn(ctx, driverID, location)
	}
	return nil
}

func (f *fakeDriverRepository) UpdateCurrentBookingID(ctx context.Context, driverID, bookingID string) error {
	if f.updateCurrentBookingIDFn != nil {
		return f.updateCurrentBookingIDFn(ctx, driverID, bookingID)
	}
	return nil
}

func (f *fakeDriverRepository) TryAssignCurrentBooking(ctx context.Context, driverID, bookingID string) (bool, error) {
	if f.tryAssignCurrentBookingFn != nil {
		return f.tryAssignCurrentBookingFn(ctx, driverID, bookingID)
	}
	return false, nil
}

func (f *fakeDriverRepository) ClearCurrentBookingIfMatches(ctx context.Context, driverID, bookingID string) error {
	if f.clearCurrentBookingFn != nil {
		return f.clearCurrentBookingFn(ctx, driverID, bookingID)
	}
	return nil
}

func (f *fakeDriverRepository) IncrementAcceptedBookings(ctx context.Context, driverID string) error {
	if f.incrementAcceptedFn != nil {
		return f.incrementAcceptedFn(ctx, driverID)
	}
	return nil
}

func (f *fakeDriverRepository) IncrementTotalBookings(ctx context.Context, driverID string) error {
	if f.incrementTotalFn != nil {
		return f.incrementTotalFn(ctx, driverID)
	}
	return nil
}

func (f *fakeDriverRepository) IncrementCompletedBookings(ctx context.Context, driverID string) error {
	if f.incrementCompletedFn != nil {
		return f.incrementCompletedFn(ctx, driverID)
	}
	return nil
}

func (f *fakeDriverRepository) GetTotalDrivers(ctx context.Context) (int64, error) {
	if f.getTotalDriversFn != nil {
		return f.getTotalDriversFn(ctx)
	}
	return 0, nil
}

type fakeVehicleRepository struct {
	createFn       func(context.Context, *models.Vehicle) error
	updateFn       func(context.Context, *models.Vehicle) error
	assignDriverFn func(context.Context, string, string) error
	deleteFn       func(context.Context, string) error
	findByIDFn     func(context.Context, string) (*models.Vehicle, error)
	findAllFn      func(context.Context) ([]*models.Vehicle, error)
}

func (f *fakeVehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	if f.createFn != nil {
		return f.createFn(ctx, vehicle)
	}
	return nil
}

func (f *fakeVehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, vehicle)
	}
	return nil
}

func (f *fakeVehicleRepository) AssignDriver(ctx context.Context, vehicleID, driverID string) error {
	if f.assignDriverFn != nil {
		return f.assignDriverFn(ctx, vehicleID, driverID)
	}
	return nil
}

func (f *fakeVehicleRepository) Delete(ctx context.Context, vehicleID string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, vehicleID)
	}
	return nil
}

func (f *fakeVehicleRepository) FindByID(ctx context.Context, vehicleID string) (*models.Vehicle, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, vehicleID)
	}
	return nil, nil
}

func (f *fakeVehicleRepository) FindAll(ctx context.Context) ([]*models.Vehicle, error) {
	if f.findAllFn != nil {
		return f.findAllFn(ctx)
	}
	return nil, nil
}

type fakeUserRepository struct {
	createFn      func(context.Context, *models.User) error
	findByEmailFn func(context.Context, string) (*models.User, error)
	findByIDFn    func(context.Context, string) (*models.User, error)
	getTotalFn    func(context.Context) (int64, error)
}

func (f *fakeUserRepository) Create(ctx context.Context, user *models.User) error {
	if f.createFn != nil {
		return f.createFn(ctx, user)
	}
	return nil
}

func (f *fakeUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if f.findByEmailFn != nil {
		return f.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (f *fakeUserRepository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeUserRepository) GetTotalUsers(ctx context.Context) (int64, error) {
	if f.getTotalFn != nil {
		return f.getTotalFn(ctx)
	}
	return 0, nil
}
