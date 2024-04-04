package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

func ConnectDB(uri string) error {
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

	return collection, nil
}

func CreateIndexes(collection *mongo.Collection) error {
ctx := context.Background()

// Create index for Age
ageIndexModel := mongo.IndexModel{
    Keys: bson.D{
        {Key: "condition.ageStart", Value: 1},
        {Key: "condition.ageEnd", Value: 1},
    },
}
_, err := collection.Indexes().CreateOne(ctx, ageIndexModel)
if err != nil {
    return err
}

// Create index for Gender
genderIndexModel := mongo.IndexModel{
    Keys: bson.D{
        {Key: "condition.gender", Value: int32(1)},
    },
}
_, err = collection.Indexes().CreateOne(ctx, genderIndexModel)
if err != nil {
    return err
}

// Create index for Country
countryIndexModel := mongo.IndexModel{
    Keys: bson.D{
        {Key: "condition.country", Value: 1},
    },
}
_, err = collection.Indexes().CreateOne(ctx, countryIndexModel)
if err != nil {
    return err
}

// Create index for Platform
platformIndexModel := mongo.IndexModel{
    Keys: bson.D{
        {Key: "condition.platform", Value: int32(1)},
    },
}
_, err = collection.Indexes().CreateOne(ctx, platformIndexModel)
if err != nil {
    return err
}

compoundIndexModel := mongo.IndexModel{
    Keys: bson.D{
        {Key: "condition.ageStart", Value: 1},
        {Key: "condition.ageEnd", Value: 1},
        {Key: "condition.gender", Value: 1},
        {Key: "condition.country", Value: 1},
        {Key: "condition.platform", Value: 1},
    },
}
_, err = collection.Indexes().CreateOne(ctx, compoundIndexModel)
if err != nil {
    return err
}

	return nil
}
