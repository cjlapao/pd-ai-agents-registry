package mongodb

import (
	"context"
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const usersCollection = "users"

func (c *Client) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	collection := c.database.Collection(usersCollection)

	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (c *Client) CreateUser(ctx context.Context, user *models.User) error {
	collection := c.database.Collection(usersCollection)

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	_, err := collection.InsertOne(ctx, user)
	return err
}
