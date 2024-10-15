package utils

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB(mongoURI string) (*mongo.Client, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, err
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        return nil, err
    }

    log.Println("Connected to MongoDB")
    return client, nil
}

func DisconnectDB(client *mongo.Client) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := client.Disconnect(ctx); err != nil {
        log.Printf("Error disconnecting from MongoDB: %v", err)
    } else {
        log.Println("Disconnected from MongoDB")
    }
}
