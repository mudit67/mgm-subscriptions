package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func MongoDbInit(uri, dbName string) (*MongoDB, error) {
	// Increased timeouts for Atlas connections
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(10).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(30 * time.Second). // Increased from 5s
		SetConnectTimeout(30 * time.Second)          // Increased from 10s

	// Create context with longer timeout for initial connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Attempting to connect to MongoDB Atlas...")

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping with retry logic
	for i := 0; i < 3; i++ {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Ping(pingCtx, nil)
		pingCancel()

		if err == nil {
			break
		}

		log.Printf("Ping attempt %d failed: %v", i+1, err)
		if i < 2 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB Atlas")

	database := client.Database(dbName)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}
