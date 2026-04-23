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
	Create(ctx context.Context, booking *models.Booking) error
	Update(ctx context.Context, booking *models.Booking) error
	FindByID(ctx context.Context, id string) (*models.Booking, error)
	AssignDriverIfUnassigned(ctx context.Context, bookingID, driverID string) (bool, error)
	FindActiveBookingByDriverID(ctx context.Context, driverID string) (*models.Booking, error)
	FindPendingScheduledBookings(ctx context.Context) ([]*models.Booking, error)
	GetActiveBookingsCount(ctx context.Context) (int64, error)
	FindAssignedBookings(ctx context.Context, driverID string) ([]*models.Booking, error)
	UpdateDriverResponseStatus(ctx context.Context, bookingID, status string) error
	GetActiveBookingsByDriverID(ctx context.Context, driverID string) ([]*models.Booking, error)
	GetActiveBookingByUserID(ctx context.Context, userID string) (*models.Booking, error)
	FindByIDAndDriverID(ctx context.Context, id string, driverID string) (*models.Booking, error)
	GetAverageTripTime(ctx context.Context) (float64, error)
	GetTotalBookings(ctx context.Context) (int64, error)
}

type bookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(dbClient *mongo.Client) BookingRepository {
	collection := dbClient.Database("logi").Collection("bookings")
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "scheduled_time", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "driver_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		utils.ErrorBackground("failed to create booking indexes", "error", err)
	}
	return &bookingRepository{collection}
}

func (r *bookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.InsertOne(opCtx, booking)
	return err
}

func (r *bookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.ReplaceOne(
		opCtx,
		bson.M{"_id": booking.ID},
		booking,
	)
	return err
}

func (r *bookingRepository) FindByID(ctx context.Context, id string) (*models.Booking, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var booking models.Booking
	err := r.collection.FindOne(opCtx, bson.M{"_id": id}).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) AssignDriverIfUnassigned(ctx context.Context, bookingID, driverID string) (bool, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	filter := bson.M{
		"_id":       bookingID,
		"driver_id": "",
		"status":    models.BookingStatusPending,
		"$or": bson.A{
			bson.M{"offered_driver_ids": driverID},
			bson.M{"offered_driver_ids": bson.M{"$exists": false}},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"driver_id":              driverID,
			"status":                 models.BookingStatusDriverAssigned,
			"driver_response_status": "Accepted",
			"offered_driver_ids":     bson.A{},
			"rejected_driver_ids":    bson.A{},
		},
	}
	result, err := r.collection.UpdateOne(opCtx, filter, update)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount == 1, nil
}

func (r *bookingRepository) FindActiveBookingByDriverID(ctx context.Context, driverID string) (*models.Booking, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var booking models.Booking
	filter := bson.M{
		"driver_id": driverID,
		"status": bson.M{
			"$in": models.DriverInProgressBookingStatuses,
		},
	}
	err := r.collection.FindOne(opCtx, filter).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) FindPendingScheduledBookings(ctx context.Context) ([]*models.Booking, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	filter := bson.M{
		"status":         models.BookingStatusPending,
		"scheduled_time": bson.M{"$lte": time.Now()},
	}
	cursor, err := r.collection.Find(opCtx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(opCtx)

	var bookings []*models.Booking
	for cursor.Next(opCtx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err != nil {
			continue
		}
		bookings = append(bookings, &booking)
	}
	return bookings, nil
}

func (r *bookingRepository) GetActiveBookingsCount(ctx context.Context) (int64, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	count, err := r.collection.CountDocuments(
		opCtx,
		bson.M{
			"status": bson.M{"$in": models.ActiveDemandBookingStatuses},
			"$or": bson.A{
				bson.M{"scheduled_time": bson.M{"$exists": false}},
				bson.M{"scheduled_time": bson.M{"$lte": time.Now()}},
			},
		},
	)
	return count, err
}

func (r *bookingRepository) UpdateDriverResponseStatus(ctx context.Context, bookingID, status string) error {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := r.collection.UpdateOne(
		opCtx,
		bson.M{"_id": bookingID},
		bson.M{"$set": bson.M{"driver_response_status": status}},
	)
	return err
}

func (r *bookingRepository) FindAssignedBookings(ctx context.Context, driverID string) ([]*models.Booking, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	filter := bson.M{
		"driver_id":              "",
		"status":                 models.BookingStatusPending,
		"offered_driver_ids":     driverID,
		"driver_response_status": "Pending",
		"rejected_driver_ids":    bson.M{"$ne": driverID},
	}
	cursor, err := r.collection.Find(opCtx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(opCtx)

	var bookings []*models.Booking
	for cursor.Next(opCtx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err != nil {
			continue
		}
		bookings = append(bookings, &booking)
	}
	return bookings, nil
}

func (r *bookingRepository) GetActiveBookingByUserID(ctx context.Context, userID string) (*models.Booking, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"status":  bson.M{"$ne": models.BookingStatusCompleted},
	}
	var booking models.Booking
	err := r.collection.FindOne(opCtx, filter).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) GetActiveBookingsByDriverID(ctx context.Context, driverID string) ([]*models.Booking, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	filter := bson.M{
		"driver_id": driverID,
		"status":    bson.M{"$ne": models.BookingStatusCompleted},
	}
	cursor, err := r.collection.Find(opCtx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(opCtx)

	var bookings []*models.Booking
	for cursor.Next(opCtx) {
		var booking models.Booking
		if err := cursor.Decode(&booking); err != nil {
			continue
		}
		bookings = append(bookings, &booking)
	}
	return bookings, nil
}

func (r *bookingRepository) FindByIDAndDriverID(ctx context.Context, id string, driverID string) (*models.Booking, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	var booking models.Booking
	filter := bson.M{
		"_id":       id,
		"driver_id": driverID,
	}
	err := r.collection.FindOne(opCtx, filter).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// GetAverageTripTime calculates the average trip time of completed bookings in minutes.
func (r *bookingRepository) GetAverageTripTime(ctx context.Context) (float64, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.D{
				{Key: "status", Value: models.BookingStatusCompleted},
				{Key: "started_at", Value: bson.D{{Key: "$exists", Value: true}}},
				{Key: "completed_at", Value: bson.D{{Key: "$exists", Value: true}}},
			},
		}},
		{{
			Key: "$project",
			Value: bson.D{
				{
					Key: "trip_time",
					Value: bson.D{{
						Key: "$divide",
						Value: bson.A{
							bson.D{{Key: "$subtract", Value: bson.A{"$completed_at", "$started_at"}}},
							1000 * 60, // Convert milliseconds to minutes
						},
					}},
				},
			},
		}},
		{{
			Key: "$group",
			Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "average_trip_time", Value: bson.D{{Key: "$avg", Value: "$trip_time"}}},
			},
		}},
	}

	cursor, err := r.collection.Aggregate(opCtx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(opCtx)

	var result struct {
		AverageTripTime float64 `bson:"average_trip_time"`
	}

	if cursor.Next(opCtx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.AverageTripTime, nil
	}

	return 0, nil // No completed bookings
}

// GetTotalBookings returns the total number of bookings.
func (r *bookingRepository) GetTotalBookings(ctx context.Context) (int64, error) {
	opCtx, cancel := utils.DBContext(ctx)
	defer cancel()

	count, err := r.collection.CountDocuments(opCtx, bson.M{})
	return count, err
}
