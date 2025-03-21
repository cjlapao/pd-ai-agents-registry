package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates all required indexes for the collections
func (c *Client) EnsureIndexes(ctx context.Context) error {
	// Package indexes
	if err := c.createPackageIndexes(ctx); err != nil {
		return err
	}

	// Version indexes
	if err := c.createVersionIndexes(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Client) createPackageIndexes(ctx context.Context) error {
	collection := c.database.Collection(packagesCollection)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (c *Client) createVersionIndexes(ctx context.Context) error {
	collection := c.database.Collection(versionsCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "package_id", Value: 1},
				{Key: "version", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}
