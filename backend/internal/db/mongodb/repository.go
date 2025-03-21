package mongodb

import (
	"context"
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	packagesCollection = "packages"
	versionsCollection = "versions"
)

// ListPackages retrieves all packages from the database
func (c *Client) ListPackages(ctx context.Context) ([]models.Package, error) {
	collection := c.database.Collection(packagesCollection)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var packages []models.Package
	if err = cursor.All(ctx, &packages); err != nil {
		return nil, err
	}

	return packages, nil
}

// GetPackage retrieves a specific package by name
func (c *Client) GetPackage(ctx context.Context, name string) (*models.Package, error) {
	collection := c.database.Collection(packagesCollection)

	var pkg models.Package
	err := collection.FindOne(ctx, bson.M{"name": name}).Decode(&pkg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &pkg, nil
}

// CreatePackage creates a new package
func (c *Client) CreatePackage(ctx context.Context, pkg *models.Package) error {
	collection := c.database.Collection(packagesCollection)

	pkg.CreatedAt = time.Now()
	pkg.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, pkg)
	return err
}

// UpdatePackage updates an existing package
func (c *Client) UpdatePackage(ctx context.Context, pkg *models.Package) error {
	collection := c.database.Collection(packagesCollection)

	pkg.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(
		ctx,
		bson.M{"_id": pkg.ID},
		pkg,
	)
	return err
}

// ListVersions retrieves all versions for a specific package
func (c *Client) ListVersions(ctx context.Context, packageID primitive.ObjectID) ([]models.Version, error) {
	collection := c.database.Collection(versionsCollection)

	cursor, err := collection.Find(ctx, bson.M{"package_id": packageID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var versions []models.Version
	if err = cursor.All(ctx, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// GetVersion retrieves a specific version of a package
func (c *Client) GetVersion(ctx context.Context, packageID primitive.ObjectID, version string) (*models.Version, error) {
	collection := c.database.Collection(versionsCollection)

	var ver models.Version
	err := collection.FindOne(ctx, bson.M{
		"package_id": packageID,
		"version":    version,
	}).Decode(&ver)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &ver, nil
}

// CreateVersion creates a new version for a package
func (c *Client) CreateVersion(ctx context.Context, version *models.Version) error {
	collection := c.database.Collection(versionsCollection)

	version.CreatedAt = time.Now()

	_, err := collection.InsertOne(ctx, version)
	return err
}

// DeleteVersion deletes a specific version of a package
func (c *Client) DeleteVersion(ctx context.Context, packageID primitive.ObjectID, version string) error {
	collection := c.database.Collection(versionsCollection)

	_, err := collection.DeleteOne(ctx, bson.M{
		"package_id": packageID,
		"version":    version,
	})
	return err
}

// AddFileToVersion adds a file to a specific version
func (c *Client) AddFileToVersion(ctx context.Context, packageID primitive.ObjectID, version string, file models.File) error {
	collection := c.database.Collection(versionsCollection)

	_, err := collection.UpdateOne(
		ctx,
		bson.M{
			"package_id": packageID,
			"version":    version,
		},
		bson.M{
			"$push": bson.M{"files": file},
		},
	)
	return err
}

// RemoveFileFromVersion removes a file from a specific version
func (c *Client) RemoveFileFromVersion(ctx context.Context, packageID primitive.ObjectID, version string, filename string) error {
	collection := c.database.Collection(versionsCollection)

	_, err := collection.UpdateOne(
		ctx,
		bson.M{
			"package_id": packageID,
			"version":    version,
		},
		bson.M{
			"$pull": bson.M{
				"files": bson.M{"name": filename},
			},
		},
	)
	return err
}
