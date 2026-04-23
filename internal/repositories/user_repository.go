package repositories

import (
	"context"
	"logi/internal/models"
	"logi/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID string) (*models.User, error)
	GetTotalUsers(ctx context.Context) (int64, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(dbClient *mongo.Client) UserRepository {
	collection := dbClient.Database("logi").Collection("users")
	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("users_email_unique"),
	})
	if err != nil {
		utils.ErrorBackground("failed to create users email index", "error", err)
	}
	return &userRepository{collection}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.InsertOne(opCtx, user)
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(opCtx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(opCtx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetTotalUsers(ctx context.Context) (int64, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	count, err := r.collection.CountDocuments(opCtx, bson.M{})
	return count, err
}
