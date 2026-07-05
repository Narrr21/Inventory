package db

import (
	"context"
	"errors"
	"os"
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

// Client is a placeholder standing in for *mongo.Client until the real
// mongo-driver dependency is introduced in the DB implementation phase.
type Client struct {
	cfg Config
}

// Connect will open the MongoDB connection once this is implemented.
// Deliberately unimplemented in this phase.
func Connect(ctx context.Context, cfg Config) (*Client, error) {
	return nil, errors.New("db: Connect not implemented yet")
}

// Disconnect will close the MongoDB connection once this is implemented.
// Deliberately unimplemented in this phase.
func Disconnect(ctx context.Context, client *Client) error {
	return errors.New("db: Disconnect not implemented yet")
}
