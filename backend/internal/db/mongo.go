package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Config holds MongoDB connection settings, sourced from environment variables.
type Config struct {
	URI      string
	Database string
}

// LoadConfig reads MONGODB_URI / MONGODB_DB from the environment.
func LoadConfig() Config {
	return Config{
		URI:      os.Getenv("MONGODB_URI"),
		Database: os.Getenv("MONGODB_DB"),
	}
}

// Client wraps a connected *mongo.Client along with the target database.
type Client struct {
	Mongo    *mongo.Client
	Database *mongo.Database
}

// Connect opens a MongoDB connection and verifies it with a ping.
func Connect(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("db: MONGODB_URI is not set")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("db: MONGODB_DB is not set")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return &Client{
		Mongo:    client,
		Database: client.Database(cfg.Database),
	}, nil
}

// Disconnect closes the MongoDB connection.
func Disconnect(ctx context.Context, client *Client) error {
	if client == nil || client.Mongo == nil {
		return nil
	}
	return client.Mongo.Disconnect(ctx)
}
