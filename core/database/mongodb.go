package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func MongoDbInit(uri, dbName string) (*MongoDB, error) {
	var client *mongo.Client
	var err error

	// Retry connection with 5-second intervals
	for {
		client, err = connectWithRetry(uri)
		if err != nil {
			log.Printf("MongoDB connection failed: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	database := client.Database(dbName)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func connectWithRetry(uri string) (*mongo.Client, error) {
	// Pool monitor to handle connection events
	poolMonitor := &event.PoolMonitor{
		Event: func(evt *event.PoolEvent) {
			switch evt.Type {
			case event.PoolClosedEvent:
				log.Println("MongoDB connection pool closed")
			case event.PoolCreated:
				log.Println("MongoDB connection pool created")
			}
		},
	}

	// Client options with retry configuration
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(10).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetHeartbeatInterval(10 * time.Second).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetPoolMonitor(poolMonitor)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Attempting to connect to MongoDB...")

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err = client.Ping(pingCtx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, err
	}

	log.Println("Successfully connected to MongoDB")
	return client, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

// Health check method
func (m *MongoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Client.Ping(ctx, nil)
}
