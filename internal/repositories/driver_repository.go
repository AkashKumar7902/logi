package repositories

import (
    "context"
    "logi/internal/models"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type DriverRepository interface {
    Create(driver *models.Driver) error
    FindByEmail(email string) (*models.Driver, error)
    FindAvailableDriver(location models.Location, vehicleType string) (*models.Driver, error)
    UpdateStatus(driverID string, status string) error
    GetAvailableDriversCount() (int64, error)
}

type driverRepository struct {
    collection *mongo.Collection
}

func NewDriverRepository(dbClient *mongo.Client) DriverRepository {
    collection := dbClient.Database("logi").Collection("drivers")
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
    // Find the nearest available driver with matching vehicle type
    filter := bson.M{
        "status":       "Available",
        "vehicle_type": vehicleType,
    }
    opts := options.FindOne()
    // For simplicity, not implementing geospatial queries
    var driver models.Driver
    err := r.collection.FindOne(context.Background(), filter, opts).Decode(&driver)
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
