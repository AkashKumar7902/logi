package repositories

import (
	"context"
	"logi/internal/models"
	"logi/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdminRepository interface {
	Create(ctx context.Context, admin *models.Admin) error
	FindByEmail(ctx context.Context, email string) (*models.Admin, error)
}

type adminRepository struct {
	collection *mongo.Collection
}

func NewAdminRepository(dbClient *mongo.Client) AdminRepository {
	collection := dbClient.Database("logi").Collection("admins")
	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("admins_email_unique"),
	})
	if err != nil {
		utils.ErrorBackground("failed to create admin email index", "error", err)
	}
	return &adminRepository{collection}
}

func (r *adminRepository) Create(ctx context.Context, admin *models.Admin) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.InsertOne(opCtx, admin)
	return err
}

func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var admin models.Admin
	err := r.collection.FindOne(opCtx, bson.M{"email": email}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}
