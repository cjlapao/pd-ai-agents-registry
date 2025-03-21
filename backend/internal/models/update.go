package models

import "time"

// Update represents an application update
type Update struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	Version     string    `bson:"version" json:"version"`
	Platform    string    `bson:"platform" json:"platform"`
	Arch        string    `bson:"arch" json:"arch"`
	Filename    string    `bson:"filename" json:"filename"`
	FileSize    int64     `bson:"file_size" json:"file_size"`
	Signature   string    `bson:"signature" json:"signature"`
	ReleaseDate time.Time `bson:"release_date" json:"release_date"`
	Notes       string    `bson:"notes" json:"notes"`
	DownloadURL string    `bson:"download_url" json:"download_url"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// UpdateMetadata represents metadata for an update
type UpdateMetadata struct {
	Version     string    `json:"version"`
	Platform    string    `json:"platform"`
	Arch        string    `json:"arch"`
	ReleaseDate time.Time `json:"release_date"`
	Notes       string    `json:"notes"`
	DownloadURL string    `json:"download_url"`
}

type LatestVersion struct {
	Version   string                           `json:"version"`
	Notes     string                           `json:"notes"`
	PubDate   string                           `json:"pub_date"`
	Platforms map[string]LatestVersionPlatform `json:"platforms"`
}

type LatestVersionPlatform struct {
	Signature string `json:"signature"`
	URL       string `json:"url"`
}
