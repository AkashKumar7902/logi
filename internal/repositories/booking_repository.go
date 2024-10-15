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
    GetBookingStatistics() (*models.BookingStatistics, error)
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

func (r *bookingRepository) GetBookingStatistics() (*models.BookingStatistics, error) {
    ctx := context.Background()
    totalBookings, err := r.collection.CountDocuments(ctx, bson.M{})
    if err != nil {
        return nil, err
    }

    completedBookings, err := r.collection.CountDocuments(ctx, bson.M{"status": "Completed"})
    if err != nil {
        return nil, err
    }

    // Calculate average trip time for completed bookings
    matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "Completed"}}}}
    projectStage := bson.D{{Key: "$project", Value: bson.D{
        {Key: "duration", Value: bson.D{{Key: "$subtract", Value: []interface{}{"$completed_at", "$started_at"}}}},
    }}}
    groupStage := bson.D{{Key: "$group", Value: bson.D{
        {Key: "_id", Value: nil},
        {Key: "averageDuration", Value: bson.D{{Key: "$avg", Value: "$duration"}}},
    }}}

    cursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{matchStage, projectStage, groupStage})
    if err != nil {
        return nil, err
    }
    var results []bson.M
    if err := cursor.All(ctx, &results); err != nil {
        return nil, err
    }

    var averageTripTime float64
    if len(results) > 0 {
        averageDuration := results[0]["averageDuration"]
        if averageDuration != nil {
            averageTripTime = float64(averageDuration.(int64)) / 60000 // Convert ms to minutes
        }
    }

    stats := &models.BookingStatistics{
        TotalBookings:     totalBookings,
        CompletedBookings: completedBookings,
        AverageTripTime:   averageTripTime,
    }

    return stats, nil
}
