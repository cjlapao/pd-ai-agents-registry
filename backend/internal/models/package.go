package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Package struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	RepoURL     string             `bson:"repo_url" json:"repo_url"`
	Author      string             `bson:"author" json:"author"`
	IsSystem    bool               `bson:"is_system" json:"is_system"`
	IsOfficial  bool               `bson:"is_official" json:"is_official"`
	Categories  []string           `bson:"categories" json:"categories"`
	Icon        string             `bson:"icon" json:"icon"`
	CompanyURL  string             `bson:"company_url" json:"company_url"`
	StarRating  float64            `bson:"star_rating" json:"star_rating"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type AgentDefinition struct {
	ID          string `bson:"id" json:"id"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Type        string `bson:"type" json:"type"`
	ModulePath  string `bson:"module_path" json:"module_path"`
	ClassName   string `bson:"class_name" json:"class_name"`
}

type Version struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PackageID    primitive.ObjectID `bson:"package_id" json:"package_id"`
	Version      string             `bson:"version" json:"version"`
	Requirements []string           `bson:"requirements" json:"requirements"`
	Agents       []AgentDefinition  `bson:"agents" json:"agents"`
	Files        []File             `bson:"files" json:"files"`
	ReleaseNotes string             `bson:"release_notes" json:"release_notes"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

type File struct {
	Name        string    `bson:"name" json:"name"`
	Size        int64     `bson:"size" json:"size"`
	Hash        string    `bson:"hash" json:"hash"`
	ContentType string    `bson:"content_type" json:"content_type"`
	DownloadURL string    `bson:"download_url" json:"download_url"`
	UploadedAt  time.Time `bson:"uploaded_at" json:"uploaded_at"`
}
