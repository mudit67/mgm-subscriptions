package database

import (
	"context"
	"log"
	"time"
)

// BackgroundReconnect monitors connection health and reconnects if needed
func (m *MongoDB) BackgroundReconnect(uri, dbName string) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.Ping(); err != nil {
				log.Printf("MongoDB ping failed: %v. Attempting reconnection...", err)
				m.reconnect(uri, dbName)
			}
		}
	}
}

func (m *MongoDB) reconnect(uri, dbName string) {
	// Close existing connection
	if m.Client != nil {
		m.Client.Disconnect(context.Background())
	}

	// Retry connection every 5 seconds
	for {
		log.Println("Attempting to reconnect to MongoDB...")

		client, err := connectWithRetry(uri)
		if err != nil {
			log.Printf("Reconnection failed: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Update client and database
		m.Client = client
		m.Database = client.Database(dbName)

		log.Println("Successfully reconnected to MongoDB")
		break
	}
}
