package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"sync"
	"time"
)

var MongoClient *mongo.Client
var mongoOnce sync.Once
var mongoDatabase string

func ConnectMongoDB() *mongo.Client {
	mongoOnce.Do(func() {
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			log.Fatal("MONGODB_URI environment variable not set")
		}
		mongoDatabase = os.Getenv("MONGODB_DATABASE")
		if mongoDatabase == "" {
			log.Fatal("MONGODB_DATABASE environment variable not set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(uri)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// Ping the database to verify the connection
		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}

		fmt.Println("Connected to MongoDB!")
		MongoClient = client

	})
	return MongoClient
}

// GetDatabase returns the MongoDB database instance. Ensure ConnectMongoDB is called first.
func GetDatabase() *mongo.Database {
	if MongoClient == nil {
		log.Fatal("MongoDB client not initialized. Call ConnectMongoDB first.")
	}
	return MongoClient.Database(mongoDatabase)
}

func CloseMongoDB() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := MongoClient.Disconnect(ctx); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
		fmt.Println("Disconnected from MongoDB.")
		MongoClient = nil
	}
}
