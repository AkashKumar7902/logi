package repositories

import (
    "context"
    "logi/internal/models"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type VehicleRepository interface {
    Create(vehicle *models.Vehicle) error
    Update(vehicle *models.Vehicle) error
    Delete(vehicleID string) error
    FindByID(vehicleID string) (*models.Vehicle, error)
    FindAll() ([]*models.Vehicle, error)
}

type vehicleRepository struct {
    collection *mongo.Collection
}

func NewVehicleRepository(dbClient *mongo.Client) VehicleRepository {
    collection := dbClient.Database("logi").Collection("vehicles")
    return &vehicleRepository{collection}
}

func (r *vehicleRepository) Create(vehicle *models.Vehicle) error {
    _, err := r.collection.InsertOne(context.Background(), vehicle)
    return err
}

func (r *vehicleRepository) Update(vehicle *models.Vehicle) error {
    _, err := r.collection.UpdateOne(
        context.Background(),
        bson.M{"_id": vehicle.ID},
        bson.M{"$set": vehicle},
    )
    return err
}

func (r *vehicleRepository) Delete(vehicleID string) error {
    _, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": vehicleID})
    return err
}

func (r *vehicleRepository) FindByID(vehicleID string) (*models.Vehicle, error) {
    var vehicle models.Vehicle
    err := r.collection.FindOne(context.Background(), bson.M{"_id": vehicleID}).Decode(&vehicle)
    if err != nil {
        return nil, err
    }
    return &vehicle, nil
}

func (r *vehicleRepository) FindAll() ([]*models.Vehicle, error) {
    cursor, err := r.collection.Find(context.Background(), bson.M{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    var vehicles []*models.Vehicle
    for cursor.Next(context.Background()) {
        var vehicle models.Vehicle
        if err := cursor.Decode(&vehicle); err != nil {
            continue
        }
        vehicles = append(vehicles, &vehicle)
    }
    return vehicles, nil
}
