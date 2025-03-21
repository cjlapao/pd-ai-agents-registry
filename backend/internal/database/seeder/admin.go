package seeder

import (
	"context"
	"fmt"
	"log"

	"github.com/Parallels/pd-ai-agents-registry/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const (
	adminEmail = "admin@example.com"
)

type AdminSeeder struct {
	db       *mongo.Database
	password string
}

func NewAdminSeeder(db *mongo.Database, password string) *AdminSeeder {
	return &AdminSeeder{
		db:       db,
		password: password,
	}
}

func (s *AdminSeeder) Seed(ctx context.Context) error {
	collection := s.db.Collection("users")

	// Check if admin user exists
	var existingUser bson.M
	err := collection.FindOne(ctx, bson.M{"email": adminEmail}).Decode(&existingUser)
	if err == nil {
		// Admin exists, check if password needs update
		if existingUser["password"] != s.password {
			log.Printf("Updating admin password")
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("error hashing password: %w", err)
			}

			_, err = collection.UpdateOne(
				ctx,
				bson.M{"email": adminEmail},
				bson.M{"$set": bson.M{"password": string(hashedPassword)}},
			)
			if err != nil {
				return fmt.Errorf("error updating admin password: %w", err)
			}
		}
		return nil
	}

	if err != mongo.ErrNoDocuments {
		return fmt.Errorf("error checking for admin user: %w", err)
	}

	// Create admin user
	log.Printf("Creating admin user")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	adminUser := models.User{
		Email:     adminEmail,
		Password:  string(hashedPassword),
		Roles:     []string{"admin"},
		Claims:    []string{"admin"},
		Username:  "admin",
		FirstName: "Admin",
		LastName:  "User",
	}

	_, err = collection.InsertOne(ctx, adminUser)
	if err != nil {
		return fmt.Errorf("error creating admin user: %w", err)
	}

	return nil
}
