package repositories

import (
    "context"
    "logi/internal/models"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
    Create(user *models.User) error
    FindByEmail(email string) (*models.User, error)
}

type userRepository struct {
    collection *mongo.Collection
}

func NewUserRepository(dbClient *mongo.Client) UserRepository {
    collection := dbClient.Database("logi").Collection("users")
    return &userRepository{collection}
}

func (r *userRepository) Create(user *models.User) error {
    _, err := r.collection.InsertOne(context.Background(), user)
    return err
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
