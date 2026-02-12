package models

import (
	"time"

	"github.com/google/uuid"
)

// FilmType represents the type of film content
type FilmType string

const (
	FilmTypeShortFilm  FilmType = "SHORT_FILM"
	FilmTypeFeatureFilm FilmType = "FEATURE_FILM"
)

// FilmStatus represents the processing status of a film
type FilmStatus string

const (
	StatusDraft      FilmStatus = "DRAFT"
	StatusUploaded   FilmStatus = "UPLOADED"
	StatusTranscoding FilmStatus = "TRANSCODING"
	StatusReady      FilmStatus = "READY"
	StatusFailed     FilmStatus = "FAILED"
)

// Film represents a video content item
type Film struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Title        string     `db:"title" json:"title"`
	Description  string     `db:"description" json:"description"`
	Duration     int        `db:"duration" json:"duration"` // in seconds
	Type         FilmType   `db:"type" json:"type"`
	Status       FilmStatus `db:"status" json:"status"`
	ThumbnailURL string     `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	HLSMasterURL string     `db:"hls_master_url" json:"hls_master_url,omitempty"`
	CreatedByID  uuid.UUID  `db:"created_by_id" json:"created_by_id"`
	CreatedBy    *User      `db:"created_by" json:"created_by,omitempty"`
	ViewCount   int        `db:"view_count" json:"view_count"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	PublishedAt *time.Time `db:"published_at" json:"published_at,omitempty"`
}

// VideoAsset represents different quality versions of a film
type VideoAsset struct {
	ID        uuid.UUID `db:"id" json:"id"`
	FilmID    uuid.UUID `db:"film_id" json:"film_id"`
	Quality   string    `db:"quality" json:"quality"` // 360p, 720p, etc.
	HLSIndexURL string   `db:"hls_index_url" json:"hls_index_url"`
	SizeBytes int64     `db:"size_bytes" json:"size_bytes"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// TranscodeJob represents a video processing job
type TranscodeJob struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	FilmID      uuid.UUID  `db:"film_id" json:"film_id"`
	Status      FilmStatus `db:"status" json:"status"`
	Error       string     `db:"error" json:"error,omitempty"`
	Progress    int        `db:"progress" json:"progress"` // 0-100
	StartedAt   *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}
