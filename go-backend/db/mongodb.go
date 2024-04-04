package db

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

func ConnectDB() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:admin@mongo:27017"))
	if err != nil {
		panic(err)
	}
    if err != nil {
        panic(err)
    }

    // Check the connection
    err = client.Ping(ctx, nil)
    if err != nil {
        panic(err)
    }

    DB = client
}
