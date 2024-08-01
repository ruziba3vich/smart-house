package storage

import (
	"context"
	"fmt"
	"log"
	"ruziba3vich/github.com/control/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	DB struct {
		Client          *mongo.Client
		UsersCollection *mongo.Collection
	}
)

func NewStorage(database *DB, logger *log.Logger) *Storage {
	return &Storage{
		database: database,
		logger:   logger,
	}
}

func ConnectDB(cfg *config.Config, ctx context.Context) (*DB, error) {
	clientOptions := options.Client().ApplyURI(cfg.DbConfig.MongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %s", err.Error())
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %s", err.Error())
	}

	return &DB{
		Client:          client,
		UsersCollection: client.Database(cfg.DbConfig.MongoDB).Collection(cfg.DbConfig.Collection),
	}, nil
}

func (db *DB) DisconnectDB(ctx context.Context) error {
	if err := db.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %s", err.Error())
	}
	return nil
}
