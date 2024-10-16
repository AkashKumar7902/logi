package repositories

import (
	"context"
	"logi/internal/models"
	"logi/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DriverRepository interface {
	Create(driver *models.Driver) error
	FindByEmail(email string) (*models.Driver, error)
	FindAvailableDriver(location models.Location, vehicleType string) (*models.Driver, error)
	UpdateStatus(driverID string, status string) error
	GetAvailableDriversCount() (int64, error)
	GetAllDrivers() ([]*models.Driver, error)
	FindByID(driverID string) (*models.Driver, error)
	UpdateDriver(driver *models.Driver) error
	UpdateLocation(driverID string, location models.Location) error
}

type driverRepository struct {
	collection *mongo.Collection
}

func NewDriverRepository(dbClient *mongo.Client) DriverRepository {
	collection := dbClient.Database("logi").Collection("drivers")

	// Create geospatial index on location
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "location", Value: "2dsphere"},
		},
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		utils.Logger.Fatalf("Failed to create geospatial index: %v", err)
	}

	return &driverRepository{collection}
}

func (r *driverRepository) Create(driver *models.Driver) error {
	_, err := r.collection.InsertOne(context.Background(), driver)
	return err
}

func (r *driverRepository) FindByEmail(email string) (*models.Driver, error) {
	var driver models.Driver
	err := r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&driver)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) FindAvailableDriver(location models.Location, vehicleType string) (*models.Driver, error) {
	filter := bson.M{
		"status":       "Available",
		"vehicle_type": vehicleType,
		"location": bson.M{
			"$near": bson.M{
				"$geometry":    location,
				"$maxDistance": 5000, // Adjust as needed (in meters)
			},
		},
	}

	var driver models.Driver
	err := r.collection.FindOne(context.Background(), filter).Decode(&driver)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) UpdateStatus(driverID string, status string) error {
	_, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": driverID},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}

func (r *driverRepository) GetAvailableDriversCount() (int64, error) {
	count, err := r.collection.CountDocuments(context.Background(), bson.M{"status": "Available"})
	return count, err
}

func (r *driverRepository) GetAllDrivers() ([]*models.Driver, error) {
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var drivers []*models.Driver
	for cursor.Next(context.Background()) {
		var driver models.Driver
		if err := cursor.Decode(&driver); err != nil {
			continue
		}
		drivers = append(drivers, &driver)
	}
	return drivers, nil
}

func (r *driverRepository) FindByID(driverID string) (*models.Driver, error) {
	var driver models.Driver
	err := r.collection.FindOne(context.Background(), bson.M{"_id": driverID}).Decode(&driver)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) UpdateDriver(driver *models.Driver) error {
	_, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": driver.ID},
		bson.M{"$set": driver},
	)
	return err
}

func (r *driverRepository) UpdateLocation(driverID string, location models.Location) error {
	_, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": driverID},
		bson.M{"$set": bson.M{
			"location": location,
		}},
	)
	return err
}
