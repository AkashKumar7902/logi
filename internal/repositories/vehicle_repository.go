package repositories

import (
	"context"
	"logi/internal/models"
	"logi/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VehicleRepository interface {
	Create(ctx context.Context, vehicle *models.Vehicle) error
	Update(ctx context.Context, vehicle *models.Vehicle) error
	Delete(ctx context.Context, vehicleID string) error
	FindByID(ctx context.Context, vehicleID string) (*models.Vehicle, error)
	FindAll(ctx context.Context) ([]*models.Vehicle, error)
}

type vehicleRepository struct {
	collection *mongo.Collection
}

func NewVehicleRepository(dbClient *mongo.Client) VehicleRepository {
	collection := dbClient.Database("logi").Collection("vehicles")
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "license_plate", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("vehicles_license_plate_unique"),
		},
		{
			Keys:    bson.D{{Key: "driver_id", Value: 1}},
			Options: options.Index().SetSparse(true).SetName("vehicles_driver_id_sparse"),
		},
	}
	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		utils.Logger.Printf("Failed to create vehicle indexes: %v", err)
	}
	return &vehicleRepository{collection}
}

func (r *vehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.InsertOne(opCtx, vehicle)
	return err
}

func (r *vehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.UpdateOne(
		opCtx,
		bson.M{"_id": vehicle.ID},
		bson.M{"$set": vehicle},
	)
	return err
}

func (r *vehicleRepository) Delete(ctx context.Context, vehicleID string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.DeleteOne(opCtx, bson.M{"_id": vehicleID})
	return err
}

func (r *vehicleRepository) FindByID(ctx context.Context, vehicleID string) (*models.Vehicle, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var vehicle models.Vehicle
	err := r.collection.FindOne(opCtx, bson.M{"_id": vehicleID}).Decode(&vehicle)
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) FindAll(ctx context.Context) ([]*models.Vehicle, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	cursor, err := r.collection.Find(opCtx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(opCtx)

	var vehicles []*models.Vehicle
	for cursor.Next(opCtx) {
		var vehicle models.Vehicle
		if err := cursor.Decode(&vehicle); err != nil {
			continue
		}
		vehicles = append(vehicles, &vehicle)
	}
	return vehicles, nil
}
