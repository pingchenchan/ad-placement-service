package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

func ConnectMongoDB(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	DB = client
	return nil
}
func DropDatabaseAndCollection(dbName, collectionName string) error {
    ctx := context.Background()

    // Drop the collection
    collection := DB.Database(dbName).Collection(collectionName)
    err := collection.Drop(ctx)
    if err != nil {
        return fmt.Errorf("failed to drop collection: %w", err)
    }

    // Drop the database
    err = DB.Database(dbName).Drop(ctx)
    if err != nil {
        return fmt.Errorf("failed to drop database: %w", err)
    }
 	log.Print("successfullly dropped database and collection")
    return nil
}
func EnsureCollectionAndIndexes(dbName, collectionName string) (*mongo.Collection, error) {
	ctx := context.Background()

	// Get a handle to the database and collection
	database := DB.Database(dbName)
	collection := database.Collection(collectionName)

	// Create the collection if it doesn't exist
	err := database.RunCommand(ctx, bson.D{{Key: "create", Value: collectionName}}).Err()
	if err != nil && !strings.Contains(err.Error(), "NamespaceExists") {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	// Create indexes
	err = CreateIndexes(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}
	log.Print("successfullly created indexes")

	return collection, nil
}

func CreateIndexes(collection *mongo.Collection) error {
	ctx := context.Background()

	compoundIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "startAt", Value: 1},
			{Key: "endAt", Value: 1},
			{Key: "condition.ageStart", Value: 1},
			{Key: "condition.ageEnd", Value: 1},
			{Key: "condition.country", Value: 1},
			{Key: "condition.gender", Value: 1},
			{Key: "condition.platform", Value: 1},
		},
	}
	_, err := collection.Indexes().CreateOne(ctx, compoundIndexModel)
	if err != nil {
		return err
	}

		return nil
}

