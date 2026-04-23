package repositories

import (
	"context"
	"logi/internal/models"
	"logi/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DriverRepository interface {
	Create(ctx context.Context, driver *models.Driver) error
	FindByEmail(ctx context.Context, email string) (*models.Driver, error)
	FindAvailableDrivers(ctx context.Context, location models.Location, vehicleType string) ([]*models.Driver, error)
	UpdateStatus(ctx context.Context, driverID string, status string) error
	AssignVehicle(ctx context.Context, driverID, vehicleID, vehicleType string) error
	GetAvailableDriversCount(ctx context.Context) (int64, error)
	GetAllDrivers(ctx context.Context) ([]*models.Driver, error)
	FindByID(ctx context.Context, driverID string) (*models.Driver, error)
	UpdateDriver(ctx context.Context, driver *models.Driver) error
	UpdateLocation(ctx context.Context, driverID string, location models.Location) error
	UpdateCurrentBookingID(ctx context.Context, driverID, bookingID string) error
	IncrementAcceptedBookings(ctx context.Context, driverID string) error
	IncrementTotalBookings(ctx context.Context, driverID string) error
	IncrementCompletedBookings(ctx context.Context, driverID string) error
	GetTotalDrivers(ctx context.Context) (int64, error)
}

type driverRepository struct {
	collection *mongo.Collection
}

func NewDriverRepository(dbClient *mongo.Client) DriverRepository {
	collection := dbClient.Database("logi").Collection("drivers")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "location", Value: "2dsphere"}},
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("drivers_email_unique"),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "vehicle_type", Value: 1},
			},
			Options: options.Index().SetName("drivers_status_vehicle_type"),
		},
		{
			Keys:    bson.D{{Key: "vehicle_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true).SetName("drivers_vehicle_id_unique"),
		},
	}
	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		utils.ErrorBackground("failed to create driver indexes", "error", err)
	}

	return &driverRepository{collection}
}

func (r *driverRepository) Create(ctx context.Context, driver *models.Driver) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	if driver.Location.Type == "" || driver.Location.Coordinates == nil {
		driver.Location = models.Location{
			Type:        "Point",
			Coordinates: []float64{0, 0}, // Default to valid values.
		}
	}
	if driver.VehicleType == "" {
		driver.VehicleType = "car"
	}
	_, err := r.collection.InsertOne(opCtx, driver)
	return err
}

func (r *driverRepository) FindByEmail(ctx context.Context, email string) (*models.Driver, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var driver models.Driver
	err := r.collection.FindOne(opCtx, bson.M{"email": email}).Decode(&driver)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) FindAvailableDrivers(ctx context.Context, location models.Location, vehicleType string) ([]*models.Driver, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	filter := bson.M{
		"status":       models.DriverStatusAvailable,
		"vehicle_type": vehicleType,
		"location": bson.M{
			"$near": bson.M{
				"$geometry":    location,
				"$maxDistance": 10000000, // Adjust as needed (in meters)
			},
		},
	}

	cursor, err := r.collection.Find(opCtx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(opCtx)

	var drivers []*models.Driver
	for cursor.Next(opCtx) {
		var driver models.Driver
		if err := cursor.Decode(&driver); err != nil {
			continue
		}
		drivers = append(drivers, &driver)
	}
	return drivers, nil
}

func (r *driverRepository) UpdateStatus(ctx context.Context, driverID string, status string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.UpdateOne(
		opCtx,
		bson.M{"_id": driverID},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}

func (r *driverRepository) AssignVehicle(ctx context.Context, driverID, vehicleID, vehicleType string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"vehicle_id":   vehicleID,
			"vehicle_type": vehicleType,
		},
	}
	_, err := r.collection.UpdateOne(opCtx, bson.M{"_id": driverID}, update)
	return err
}

func (r *driverRepository) GetAvailableDriversCount(ctx context.Context) (int64, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	count, err := r.collection.CountDocuments(opCtx, bson.M{"status": models.DriverStatusAvailable})
	return count, err
}

func (r *driverRepository) GetAllDrivers(ctx context.Context) ([]*models.Driver, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	cursor, err := r.collection.Find(opCtx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(opCtx)

	var drivers []*models.Driver
	for cursor.Next(opCtx) {
		var driver models.Driver
		if err := cursor.Decode(&driver); err != nil {
			continue
		}
		drivers = append(drivers, &driver)
	}
	return drivers, nil
}

func (r *driverRepository) FindByID(ctx context.Context, driverID string) (*models.Driver, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var driver models.Driver
	err := r.collection.FindOne(opCtx, bson.M{"_id": driverID}).Decode(&driver)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) UpdateDriver(ctx context.Context, driver *models.Driver) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.ReplaceOne(
		opCtx,
		bson.M{"_id": driver.ID},
		driver,
	)
	return err
}

func (r *driverRepository) UpdateLocation(ctx context.Context, driverID string, location models.Location) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.UpdateOne(
		opCtx,
		bson.M{"_id": driverID},
		bson.M{"$set": bson.M{
			"location": location,
		}},
	)
	return err
}

func (r *driverRepository) UpdateCurrentBookingID(ctx context.Context, driverID, bookingID string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.UpdateOne(
		opCtx,
		bson.M{"_id": driverID},
		bson.M{"$set": bson.M{"current_booking_id": bookingID}},
	)
	return err
}

func (r *driverRepository) IncrementAcceptedBookings(ctx context.Context, driverID string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	update := bson.M{
		"$inc": bson.M{
			"accepted_bookings_count": 1,
		},
	}
	_, err := r.collection.UpdateOne(opCtx, bson.M{"_id": driverID}, update)
	return err
}

func (r *driverRepository) IncrementTotalBookings(ctx context.Context, driverID string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	update := bson.M{
		"$inc": bson.M{
			"total_bookings_count": 1,
		},
	}
	_, err := r.collection.UpdateOne(opCtx, bson.M{"_id": driverID}, update)
	return err
}

func (r *driverRepository) IncrementCompletedBookings(ctx context.Context, driverID string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	update := bson.M{
		"$inc": bson.M{
			"completed_bookings_count": 1,
		},
	}
	_, err := r.collection.UpdateOne(opCtx, bson.M{"_id": driverID}, update)
	return err
}

// GetTotalDrivers returns the total number of drivers.
func (r *driverRepository) GetTotalDrivers(ctx context.Context) (int64, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	count, err := r.collection.CountDocuments(opCtx, bson.M{})
	return count, err
}
