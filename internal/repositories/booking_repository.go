package repositories

import (
	"context"
	"logi/internal/models"
	"logi/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository interface {
    Create(booking *models.Booking) error
    Update(booking *models.Booking) error
    FindByID(id string) (*models.Booking, error)
    FindActiveBookingByDriverID(driverID string) (*models.Booking, error)
    FindPendingScheduledBookings() ([]*models.Booking, error)
    GetActiveBookingsCount() (int64, error)
    FindAssignedBookings(driverID string) ([]*models.Booking, error)
    UpdateDriverResponseStatus(bookingID, status string) error
    GetActiveBookingsByDriverID(driverID string) ([]*models.Booking, error)
    GetActiveBookingByUserID(userID string) (*models.Booking, error)
    FindByIDAndDriverID(id string, driverID string) (*models.Booking, error)
    GetAverageTripTime() (float64, error)
    GetTotalBookings() (int64, error)
}

type bookingRepository struct {
    collection *mongo.Collection
}

func NewBookingRepository(dbClient *mongo.Client) BookingRepository {
    collection := dbClient.Database("logi").Collection("bookings")
    return &bookingRepository{collection}
}

func (r *bookingRepository) Create(booking *models.Booking) error {
    utils.Logger.Println("creating booking")
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

func (r *bookingRepository) FindActiveBookingByDriverID(driverID string) (*models.Booking, error) {
    var booking models.Booking
    filter := bson.M{
        "driver_id": driverID,
        "status": bson.M{
            "$in": []string{
                "Driver Assigned",
                "En Route to Pickup",
                "Goods Collected",
                "In Transit",
            },
        },
    }
    err := r.collection.FindOne(context.Background(), filter).Decode(&booking)
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

func (r *bookingRepository) UpdateDriverResponseStatus(bookingID, status string) error {
    _, err := r.collection.UpdateOne(
        context.Background(),
        bson.M{"_id": bookingID},
        bson.M{"$set": bson.M{"driver_response_status": status}},
    )
    return err
}

func (r *bookingRepository) FindAssignedBookings(driverID string) ([]*models.Booking, error) {
    filter := bson.M{
        "driver_id": driverID,
        "driver_response_status": "Pending",
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

func (r *bookingRepository) GetActiveBookingByUserID(userID string) (*models.Booking, error) {
    filter := bson.M{
        "user_id": userID,
        "status": bson.M{
            "$nin": []string{"Completed", "Pending"},
        },
    }
    var booking models.Booking
    err := r.collection.FindOne(context.Background(), filter).Decode(&booking)
    if err != nil {
        return nil, err
    }
    return &booking, nil
}

func (r *bookingRepository) GetActiveBookingsByDriverID(driverID string) ([]*models.Booking, error) {
    filter := bson.M{
        "driver_id": driverID,
        "status": bson.M{
            "$nin": []string{"Completed", "Pending"},
        },
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

func (r *bookingRepository) FindByIDAndDriverID(id string, driverID string) (*models.Booking, error) {
    var booking models.Booking
    filter := bson.M{
        "_id":       id,
        "driver_id": driverID,
    }
    err := r.collection.FindOne(context.Background(), filter).Decode(&booking)
    if err != nil {
        return nil, err
    }
    return &booking, nil
}

// GetAverageTripTime calculates the average trip time of completed bookings in minutes.
func (r *bookingRepository) GetAverageTripTime() (float64, error) {
    pipeline := mongo.Pipeline{
        {{"$match", bson.D{
            {"status", "Completed"},
            {"started_at", bson.D{{"$exists", true}}},
            {"completed_at", bson.D{{"$exists", true}}},
        }}},
        {{"$project", bson.D{
            {"trip_time", bson.D{
                {"$divide", bson.A{
                    bson.D{{"$subtract", bson.A{"$completed_at", "$started_at"}}},
                    1000 * 60, // Convert milliseconds to minutes
                }},
            }},
        }}},
        {{"$group", bson.D{
            {"_id", nil},
            {"average_trip_time", bson.D{{"$avg", "$trip_time"}}},
        }}},
    }

    cursor, err := r.collection.Aggregate(context.Background(), pipeline)
    if err != nil {
        return 0, err
    }
    defer cursor.Close(context.Background())

    var result struct {
        AverageTripTime float64 `bson:"average_trip_time"`
    }

    if cursor.Next(context.Background()) {
        if err := cursor.Decode(&result); err != nil {
            return 0, err
        }
        return result.AverageTripTime, nil
    }

    return 0, nil // No completed bookings
}

// GetTotalBookings returns the total number of bookings.
func (r *bookingRepository) GetTotalBookings() (int64, error) {
    count, err := r.collection.CountDocuments(context.Background(), bson.M{})
    return count, err
}
