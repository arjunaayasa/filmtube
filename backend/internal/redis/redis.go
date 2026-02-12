package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arjunaayasa/filmtube/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
)

const (
	// Queue names
	TranscodeQueue = "filmtube:transcode:queue"

	// Key patterns
	TranscodeJobKey = "filmtube:transcode:job:%s"
	FilmStatusKey   = "filmtube:film:status:%s"
)

type Client struct {
	*redis.Client
}

// New creates a new Redis client
func New(addr, password string, db int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{Client: rdb}, nil
}

// ========== TRANSCODE QUEUE OPERATIONS ==========

// EnqueueTranscodeJob adds a film ID to the transcode queue
func (c *Client) EnqueueTranscodeJob(ctx context.Context, filmID uuid.UUID) error {
	return c.LPush(ctx, TranscodeQueue, filmID.String()).Err()
}

// DequeueTranscodeJob removes and returns a film ID from the queue (blocking)
func (c *Client) DequeueTranscodeJob(ctx context.Context, timeout time.Duration) (uuid.UUID, error) {
	result, err := c.BRPop(ctx, timeout, TranscodeQueue).Result()
	if err != nil {
		return uuid.Nil, err
	}

	filmID, err := uuid.Parse(result[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid film ID in queue: %w", err)
	}

	return filmID, nil
}

// SetTranscodeJobProgress stores job progress in Redis
func (c *Client) SetTranscodeJobProgress(ctx context.Context, filmID uuid.UUID, job *models.TranscodeJob) error {
	key := fmt.Sprintf(TranscodeJobKey, filmID)
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return c.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetTranscodeJobProgress retrieves job progress from Redis
func (c *Client) GetTranscodeJobProgress(ctx context.Context, filmID uuid.UUID) (*models.TranscodeJob, error) {
	key := fmt.Sprintf(TranscodeJobKey, filmID)
	data, err := c.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var job models.TranscodeJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// SetFilmStatus caches film status in Redis
func (c *Client) SetFilmStatus(ctx context.Context, filmID uuid.UUID, status models.FilmStatus) error {
	key := fmt.Sprintf(FilmStatusKey, filmID)
	return c.Set(ctx, key, string(status), 5*time.Minute).Err()
}

// GetFilmStatus retrieves cached film status from Redis
func (c *Client) GetFilmStatus(ctx context.Context, filmID uuid.UUID) (models.FilmStatus, error) {
	key := fmt.Sprintf(FilmStatusKey, filmID)
	result, err := c.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return models.FilmStatus(result), nil
}
