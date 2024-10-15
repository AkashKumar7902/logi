package repositories

import (
    "context"
    "logi/internal/models"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type AdminRepository interface {
    Create(admin *models.Admin) error
    FindByEmail(email string) (*models.Admin, error)
}

type adminRepository struct {
    collection *mongo.Collection
}

func NewAdminRepository(dbClient *mongo.Client) AdminRepository {
    collection := dbClient.Database("logi").Collection("admins")
    return &adminRepository{collection}
}

func (r *adminRepository) Create(admin *models.Admin) error {
    _, err := r.collection.InsertOne(context.Background(), admin)
    return err
}

func (r *adminRepository) FindByEmail(email string) (*models.Admin, error) {
    var admin models.Admin
    err := r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&admin)
    if err != nil {
        return nil, err
    }
    return &admin, nil
}
