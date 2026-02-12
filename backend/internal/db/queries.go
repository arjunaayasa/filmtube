package db

import (
	"context"
	"time"

	"github.com/arjunaayasa/filmtube/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/google/uuid"
)

// Queries contains all database operations
type Queries struct {
	db *DB
}

// NewQueries creates a new Queries instance
func NewQueries(db *DB) *Queries {
	return &Queries{db: db}
}

// ========== USER QUERIES ==========

// CreateUser inserts a new user
func (q *Queries) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, name, avatar_url, bio)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := q.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Role,
		user.Name, user.AvatarURL, user.Bio,
	)
	return err
}

// GetUserByID retrieves a user by ID
func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE id = $1`
	err := q.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE email = $1`
	err := q.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ========== FILM QUERIES ==========

// CreateFilm inserts a new film
func (q *Queries) CreateFilm(ctx context.Context, film *models.Film) error {
	query := `
		INSERT INTO films (id, title, description, duration, type, status, created_by_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`
	rows, err := q.db.QueryxContext(ctx, query,
		film.ID, film.Title, film.Description, film.Duration,
		film.Type, film.Status, film.CreatedByID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.StructScan(film)
}

// GetFilmByID retrieves a film by ID
func (q *Queries) GetFilmByID(ctx context.Context, id uuid.UUID) (*models.Film, error) {
	var film models.Film
	query := `
		SELECT f.*,
		       COALESCE(jsonb_build_object(
		           'id', u.id,
		           'email', u.email,
		           'name', u.name,
		           'avatar_url', u.avatar_url
		       )::json, '{}'::json) as created_by
		FROM films f
		LEFT JOIN users u ON f.created_by_id = u.id
		WHERE f.id = $1
	`
	err := q.db.GetContext(ctx, &film, query, id)
	if err != nil {
		return nil, err
	}
	return &film, nil
}

// ListFilms retrieves films with pagination
func (q *Queries) ListFilms(ctx context.Context, limit int, offset int, status models.FilmStatus) ([]models.Film, error) {
	var films []models.Film
	query := `
		SELECT f.*,
		       COALESCE(jsonb_build_object(
		           'id', u.id,
		           'email', u.email,
		           'name', u.name,
		           'avatar_url', u.avatar_url
		       )::json, '{}'::json) as created_by
		FROM films f
		LEFT JOIN users u ON f.created_by_id = u.id
		WHERE ($1 = '' OR status = $1)
		ORDER BY published_at DESC NULLS LAST, created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := q.db.SelectContext(ctx, &films, query, status, limit, offset)
	return films, err
}

// UpdateFilmStatus updates the status of a film
func (q *Queries) UpdateFilmStatus(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status models.FilmStatus) error {
	query := `UPDATE films SET status = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, status, id)
	return err
}

// UpdateFilmHLS updates HLS URLs for a film
func (q *Queries) UpdateFilmHLS(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, masterURL, thumbnailURL string) error {
	query := `
		UPDATE films
		SET hls_master_url = $1,
		    thumbnail_url = $2,
		    status = 'READY'
		WHERE id = $3
	`
	_, err := tx.ExecContext(ctx, query, masterURL, thumbnailURL, id)
	return err
}

// PublishFilm publishes a film (sets published_at and status to READY)
func (q *Queries) PublishFilm(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error {
	query := `
		UPDATE films
		SET published_at = NOW(),
		    status = 'READY'
		WHERE id = $1 AND status = 'DRAFT'
	`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

// IncrementViewCount increments the view count for a film
func (q *Queries) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE films SET view_count = view_count + 1 WHERE id = $1`
	_, err := q.db.ExecContext(ctx, query, id)
	return err
}

// ========== TRANSCODE JOB QUERIES ==========

// CreateTranscodeJob creates a new transcode job
func (q *Queries) CreateTranscodeJob(ctx context.Context, job *models.TranscodeJob) error {
	query := `
		INSERT INTO transcode_jobs (id, film_id, status, progress)
		VALUES ($1, $2, $3, $4)
	`
	_, err := q.db.ExecContext(ctx, query,
		job.ID, job.FilmID, job.Status, job.Progress,
	)
	return err
}

// GetNextTranscodeJob retrieves the next pending job
func (q *Queries) GetNextTranscodeJob(ctx context.Context) (*models.TranscodeJob, error) {
	var job models.TranscodeJob
	query := `
		SELECT * FROM transcode_jobs
		WHERE status IN ('UPLOADED', 'TRANSCODING')
		ORDER BY created_at ASC
		LIMIT 1
	`
	err := q.db.GetContext(ctx, &job, query)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateTranscodeJobStatus updates job status and progress
func (q *Queries) UpdateTranscodeJobStatus(ctx context.Context, id uuid.UUID, status models.FilmStatus, progress int, errorMsg string) error {
	query := `
		UPDATE transcode_jobs
		SET status = $1,
		    progress = $2,
		    error = $3,
		    started_at = CASE WHEN $4 AND started_at IS NULL THEN NOW() ELSE started_at END,
		    completed_at = CASE WHEN $5 THEN NOW() ELSE completed_at END
		WHERE id = $6
	`
	isStarted := status == models.StatusTranscoding
	isCompleted := status == models.StatusReady || status == models.StatusFailed
	_, err := q.db.ExecContext(ctx, query, status, progress, errorMsg, isStarted, isCompleted, id)
	return err
}

// ========== VIDEO ASSET QUERIES ==========

// CreateVideoAsset inserts a new video asset
func (q *Queries) CreateVideoAsset(ctx context.Context, asset *models.VideoAsset) error {
	query := `
		INSERT INTO video_assets (id, film_id, quality, hls_index_url, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (film_id, quality) DO UPDATE
		SET hls_index_url = EXCLUDED.hls_index_url,
		    size_bytes = EXCLUDED.size_bytes
	`
	_, err := q.db.ExecContext(ctx, query,
		asset.ID, asset.FilmID, asset.Quality,
		asset.HLSIndexURL, asset.SizeBytes,
	)
	return err
}

// GetVideoAssetsByFilmID retrieves all video assets for a film
func (q *Queries) GetVideoAssetsByFilmID(ctx context.Context, filmID uuid.UUID) ([]models.VideoAsset, error) {
	var assets []models.VideoAsset
	query := `SELECT * FROM video_assets WHERE film_id = $1 ORDER BY quality DESC`
	err := q.db.SelectContext(ctx, &assets, query, filmID)
	return assets, err
}
