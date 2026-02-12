package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arjunaayasa/filmtube/backend/internal/db"
	"github.com/arjunaayasa/filmtube/backend/internal/r2"
	"github.com/arjunaayasa/filmtube/backend/internal/redis"
	"github.com/arjunaayasa/filmtube/worker/internal/config"
	"github.com/arjunaayasa/filmtube/worker/internal/ffmpeg"
	"github.com/arjunaayasa/filmtube/worker/internal/jobs"
	"github.com/google/uuid"
)

func main() {
	log.Println("FilmTube Transcoding Worker starting...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize Redis
	redisClient, err := redis.New(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize R2 client
	r2Client, err := r2.New(
		cfg.R2Endpoint,
		cfg.R2AccessKeyID,
		cfg.R2SecretAccessKey,
		cfg.R2Bucket,
		cfg.R2Region,
		cfg.R2PublicURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}

	// Initialize FFmpeg handler
	ffmpegHandler := ffmpeg.New(cfg.FFmpegPath, cfg.TempDir)

	// Initialize processor
	queries := db.NewQueries(database)
	processor := jobs.NewProcessor(queries, r2Client, redisClient, ffmpegHandler)

	// Start worker loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workerLoop(ctx, processor, redisClient)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker shutting down...")
	cancel()
	time.Sleep(2 * time.Second)
	log.Println("Worker stopped")
}

// workerLoop continuously polls for and processes transcoding jobs
func workerLoop(ctx context.Context, processor *jobs.Processor, redisClient *redis.Client) {
	log.Println("Worker loop started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker loop stopped")
			return

		default:
			// Try to dequeue a job (with 5 second timeout)
			filmID, err := redisClient.DequeueTranscodeJob(ctx, 5*time.Second)
			if err != nil {
				if err.Error() != "redis: nil" {
					log.Printf("Error dequeuing job: %v", err)
				}
				continue
			}

			if filmID == uuid.Nil {
				continue
			}

			log.Printf("Received job for film: %s", filmID)

			// Process the job
			if err := processor.ProcessJob(ctx, filmID); err != nil {
				log.Printf("Error processing job for film %s: %v", filmID, err)
			}
		}
	}
}
