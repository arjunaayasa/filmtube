package jobs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arjunaayasa/filmtube/backend/internal/db"
	"github.com/arjunaayasa/filmtube/backend/internal/models"
	"github.com/arjunaayasa/filmtube/backend/internal/r2"
	"github.com/arjunaayasa/filmtube/backend/internal/redis"
	"github.com/arjunaayasa/filmtube/worker/internal/ffmpeg"
	"github.com/google/uuid"
)

// Processor handles video transcoding jobs
type Processor struct {
	queries   *db.Queries
	r2Client  *r2.Client
	redis     *redis.Client
	ffmpeg    *ffmpeg.FFmpeg
}

func NewProcessor(queries *db.Queries, r2Client *r2.Client, redisClient *redis.Client, ffmpeg *ffmpeg.FFmpeg) *Processor {
	return &Processor{
		queries:  queries,
		r2Client: r2Client,
		redis:    redisClient,
		ffmpeg:   ffmpeg,
	}
}

// ProcessJob processes a single transcoding job for a film
func (p *Processor) ProcessJob(ctx context.Context, filmID uuid.UUID) error {
	log.Printf("[Job] Starting transcoding for film %s", filmID)

	// Update job status to TRANSCODING
	if err := p.queries.UpdateTranscodeJobStatus(ctx, filmID, models.StatusTranscoding, 10, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Download original video from R2
	log.Printf("[Job] Downloading video from R2...")
	videoData, err := p.r2Client.DownloadOriginalVideo(ctx, filmID)
	if err != nil {
		p.markFailed(ctx, filmID, fmt.Sprintf("failed to download video: %v", err))
		return fmt.Errorf("failed to download video: %w", err)
	}

	// Get video info
	log.Printf("[Job] Getting video info...")
	ffmpegHandler := ffmpeg.New("ffmpeg", "/tmp")
	videoInfo, err := ffmpegHandler.GetVideoInfo(videoData)
	if err != nil {
		p.markFailed(ctx, filmID, fmt.Sprintf("failed to get video info: %v", err))
		return fmt.Errorf("failed to get video info: %w", err)
	}

	log.Printf("[Job] Video info: duration=%v, resolution=%dx%d",
		videoInfo.Duration, videoInfo.Width, videoInfo.Height)

	// Update progress
	p.queries.UpdateTranscodeJobStatus(ctx, filmID, models.StatusTranscoding, 20, "")

	// Generate thumbnail at 10% of video
	thumbnailTime := time.Duration(float64(videoInfo.Duration) * 0.1)
	thumbnailData, err := ffmpegHandler.GenerateThumbnail(videoData, thumbnailTime)
	if err != nil {
		log.Printf("[Job] Warning: failed to generate thumbnail: %v", err)
	} else {
		// Upload thumbnail to R2
		thumbnailKey := fmt.Sprintf("%s/%s/poster.jpg", r2.ThumbnailPath, filmID)
		if err := p.r2Client.UploadFile(ctx, thumbnailKey, bytes.NewReader(thumbnailData), "image/jpeg"); err != nil {
			log.Printf("[Job] Warning: failed to upload thumbnail: %v", err)
		}
	}

	// Transcode to each quality
	completedQualities := []string{}
	progressChan := make(chan int, 100)

	for i, quality := range ffmpeg.Qualities {
		log.Printf("[Job] Transcoding to %s...", quality.Name)

		// Start transcoding
		resultChan := make(chan *ffmpeg.TranscodeResult, 1)
		errChan := make(chan error, 1)

		go func(q ffmpeg.QualityLevel) {
			result, err := ffmpegHandler.TranscodeToHLS(videoData, filmID.String(), q, progressChan)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- result
		}(quality)

		// Wait for result
		select {
		case err := <-errChan:
			p.markFailed(ctx, filmID, fmt.Sprintf("failed to transcode to %s: %v", quality.Name, err))
			return fmt.Errorf("transcoding failed for %s: %w", quality.Name, err)

		case result := <-resultChan:
			// Upload HLS files to R2
			log.Printf("[Job] Uploading HLS files for %s...", quality.Name)
			if err := p.uploadHLSFiles(ctx, filmID, quality.Name, result.IndexData); err != nil {
				p.markFailed(ctx, filmID, fmt.Sprintf("failed to upload HLS files: %v", err))
				return fmt.Errorf("failed to upload HLS files: %w", err)
			}
			completedQualities = append(completedQualities, quality.Name)
		}

		// Update progress (20-80% for transcoding)
		baseProgress := 20
		progressPerQuality := 60 / len(ffmpeg.Qualities)
		currentProgress := baseProgress + (i+1)*progressPerQuality
		p.queries.UpdateTranscodeJobStatus(ctx, filmID, models.StatusTranscoding, currentProgress, "")
	}

	// Generate and upload master playlist
	log.Printf("[Job] Generating master playlist...")
	masterData, err := ffmpegHandler.GenerateMasterPlaylist(filmID.String(), completedQualities)
	if err != nil {
		p.markFailed(ctx, filmID, fmt.Sprintf("failed to generate master playlist: %v", err))
		return fmt.Errorf("failed to generate master playlist: %w", err)
	}

	// Upload master playlist
	masterKey := fmt.Sprintf("%s/%s/master.m3u8", r2.HLSPath, filmID)
	if err := p.r2Client.UploadFile(ctx, masterKey, bytes.NewReader(masterData), "application/x-mpegURL"); err != nil {
		p.markFailed(ctx, filmID, fmt.Sprintf("failed to upload master playlist: %v", err))
		return fmt.Errorf("failed to upload master playlist: %w", err)
	}

	// Update film status to READY
	log.Printf("[Job] Updating film status to READY...")
	tx, _ := p.queries.db.BeginTx(ctx, nil)
	masterURL := p.r2Client.GetHLSMasterURL(filmID)
	thumbnailURL := p.r2Client.GetThumbnailURL(filmID)
	if err := p.queries.UpdateFilmHLS(ctx, tx, filmID, masterURL, thumbnailURL); err != nil {
		tx.Rollback()
		p.markFailed(ctx, filmID, fmt.Sprintf("failed to update film: %v", err))
		return fmt.Errorf("failed to update film: %w", err)
	}
	tx.Commit()

	// Mark job as complete
	p.queries.UpdateTranscodeJobStatus(ctx, filmID, models.StatusReady, 100, "")

	// Update Redis cache
	p.redis.SetFilmStatus(ctx, filmID, models.StatusReady)

	log.Printf("[Job] Transcoding completed successfully for film %s", filmID)
	return nil
}

func (p *Processor) uploadHLSFiles(ctx context.Context, filmID uuid.UUID, quality string, indexData []byte) error {
	// Upload index.m3u8
	if err := p.r2Client.UploadHLSFile(ctx, filmID, quality, "index.m3u8", bytes.NewReader(indexData)); err != nil {
		return err
	}

	// TODO: In a real implementation, you would upload all .ts segments here
	// For this MVP, we're assuming segments are handled inline

	return nil
}

func (p *Processor) markFailed(ctx context.Context, filmID uuid.UUID, errorMsg string) {
	log.Printf("[Job] Marking job as failed: %s", errorMsg)
	p.queries.UpdateTranscodeJobStatus(ctx, filmID, models.StatusFailed, 0, errorMsg)
	p.redis.SetFilmStatus(ctx, filmID, models.StatusFailed)

	// Also update film status to FAILED
	tx, _ := p.queries.db.BeginTx(ctx, nil)
	p.queries.UpdateFilmStatus(ctx, tx, filmID, models.StatusFailed)
	tx.Commit()
}
