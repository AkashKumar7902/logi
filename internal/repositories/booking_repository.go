package repositories

import (
    "context"
    "logi/internal/models"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository interface {
    Create(booking *models.Booking) error
    Update(booking *models.Booking) error
    FindByID(id string) (*models.Booking, error)
    FindPendingScheduledBookings() ([]*models.Booking, error)
    GetActiveBookingsCount() (int64, error)
}

type bookingRepository struct {
    collection *mongo.Collection
}

func NewBookingRepository(dbClient *mongo.Client) BookingRepository {
    collection := dbClient.Database("logi").Collection("bookings")
    return &bookingRepository{collection}
}

func (r *bookingRepository) Create(booking *models.Booking) error {
    _, err := r.collection.InsertOne(context.Background(), booking)
    return err
}

func (r *bookingRepository) Update(booking *models.Booking) error {
    _, err := r.collection.UpdateOne(
        context.Background(),
        bson.M{"_id": booking.ID},
        bson.M{"$set": booking},
    )
    return err
}

func (r *bookingRepository) FindByID(id string) (*models.Booking, error) {
    var booking models.Booking
    err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&booking)
    if err != nil {
        return nil, err
    }
    return &booking, nil
}

func (r *bookingRepository) FindPendingScheduledBookings() ([]*models.Booking, error) {
    filter := bson.M{
        "status":        "Pending",
        "scheduled_time": bson.M{"$lte": time.Now()},
    }
    cursor, err := r.collection.Find(context.Background(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    var bookings []*models.Booking
    for cursor.Next(context.Background()) {
        var booking models.Booking
        if err := cursor.Decode(&booking); err != nil {
            continue
        }
        bookings = append(bookings, &booking)
    }
    return bookings, nil
}

func (r *bookingRepository) GetActiveBookingsCount() (int64, error) {
    count, err := r.collection.CountDocuments(
        context.Background(),
        bson.M{"status": bson.M{"$in": []string{"Pending", "Driver Assigned", "In Transit"}}},
    )
    return count, err
}
