package seeder

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

type Seeder interface {
	Seed(ctx context.Context) error
}

type DatabaseSeeder struct {
	db      *mongo.Database
	seeders []Seeder
}

func NewDatabaseSeeder(db *mongo.Database, adminPassword string) *DatabaseSeeder {
	return &DatabaseSeeder{
		db: db,
		seeders: []Seeder{
			NewAdminSeeder(db, adminPassword),
			// Add more seeders here as needed
		},
	}
}

func (s *DatabaseSeeder) Seed(ctx context.Context) error {
	log.Printf("Starting database seeding...")
	for _, seeder := range s.seeders {
		if err := seeder.Seed(ctx); err != nil {
			return fmt.Errorf("error running seeder: %w", err)
		}
	}
	log.Printf("Database seeding completed successfully")
	return nil
}
