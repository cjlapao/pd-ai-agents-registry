package mongodb

import (
	"context"
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewClient(ctx context.Context, cfg config.MongoDBConfig) (*Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Client{
		client:   client,
		database: client.Database(cfg.Database),
	}, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) Database() *mongo.Database {
	return c.database
}
