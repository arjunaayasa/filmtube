package api

import (
	"net/http"
	"strconv"

	"github.com/arjunaayasa/filmtube/internal/db"
	"github.com/arjunaayasa/filmtube/internal/models"
	"github.com/arjunaayasa/filmtube/internal/r2"
	"github.com/arjunaayasa/filmtube/internal/redis"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FilmHandler handles film endpoints
type FilmHandler struct {
	queries    *db.Queries
	r2Client   *r2.Client
	redis      *redis.Client
	expiration int // minutes for upload URLs
}

func NewFilmHandler(queries *db.Queries, r2Client *r2.Client, redisClient *redis.Client, uploadExpirationMinutes int) *FilmHandler {
	return &FilmHandler{
		queries:    queries,
		r2Client:   r2Client,
		redis:      redisClient,
		expiration: uploadExpirationMinutes,
	}
}

// CreateFilmRequest represents film creation input
type CreateFilmRequest struct {
	Title       string `json:"title" binding:"required,max=500"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required,oneof=SHORT_FILM FEATURE_FILM"`
}

// UpdateFilmRequest represents film update input
type UpdateFilmRequest struct {
	Title       string `json:"title" binding:"required,max=500"`
	Description string `json:"description"`
}

// CreateFilm creates a new film
func (h *FilmHandler) CreateFilm(c *gin.Context) {
	var req CreateFilmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := GetUserID(c)

	film := &models.Film{
		ID:           uuid.New(),
		Title:        req.Title,
		Description:  req.Description,
		Type:         models.FilmType(req.Type),
		Status:       models.StatusDraft,
		CreatedByID:  userID,
	}

	if err := h.queries.CreateFilm(c.Request.Context(), film); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create film"})
		return
	}

	c.JSON(http.StatusCreated, film)
}

// GetFilm retrieves a film by ID
func (h *FilmHandler) GetFilm(c *gin.Context) {
	idParam := c.Param("id")
	filmID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid film ID"})
		return
	}

	film, err := h.queries.GetFilmByID(c.Request.Context(), filmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "film not found"})
		return
	}

	c.JSON(http.StatusOK, film)
}

// ListFilms retrieves films with pagination
func (h *FilmHandler) ListFilms(c *gin.Context) {
	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	statusStr := c.DefaultQuery("status", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	var status models.FilmStatus
	if statusStr == "READY" {
		status = models.StatusReady
	} else {
		status = ""
	}

	films, err := h.queries.ListFilms(c.Request.Context(), limit, offset, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve films"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"films": films,
		"page":  page,
		"limit": limit,
	})
}

// GetUploadURL generates a pre-signed URL for video upload
func (h *FilmHandler) GetUploadURL(c *gin.Context) {
	idParam := c.Param("id")
	filmID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid film ID"})
		return
	}

	ctx := c.Request.Context()

	// Get film to verify ownership
	film, err := h.queries.GetFilmByID(ctx, filmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "film not found"})
		return
	}

	// Check ownership
	userID, _ := GetUserID(c)
	if film.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to upload to this film"})
		return
	}

	// Generate upload URL
	expiration := h.redis.Client.Options().ReadTimeout
	if expiration == 0 {
		expiration = 30 * 60 // 30 minutes default
	}

	uploadURL, err := h.r2Client.GeneratePresignedUploadURL(ctx, filmID, expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload URL"})
		return
	}

	// Update film status to UPLOADED (in transaction)
	tx, err := h.queries.db.BeginTx(ctx, nil)
	if err == nil {
		h.queries.UpdateFilmStatus(ctx, tx, filmID, models.StatusUploaded)
		tx.Commit()
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url":    uploadURL,
		"expiration":    expiration.String(),
		"max_file_size": 2147483648, // 2GB in bytes
	})
}

// ConfirmUpload is called after successful upload to trigger transcoding
func (h *FilmHandler) ConfirmUpload(c *gin.Context) {
	idParam := c.Param("id")
	filmID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid film ID"})
		return
	}

	ctx := c.Request.Context()

	// Get film to verify ownership
	film, err := h.queries.GetFilmByID(ctx, filmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "film not found"})
		return
	}

	// Check ownership
	userID, _ := GetUserID(c)
	if film.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	// Create transcode job
	job := &models.TranscodeJob{
		ID:       uuid.New(),
		FilmID:   filmID,
		Status:   models.StatusUploaded,
		Progress: 0,
	}

	if err := h.queries.CreateTranscodeJob(ctx, job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transcode job"})
		return
	}

	// Enqueue job for worker
	if err := h.redis.EnqueueTranscodeJob(ctx, filmID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue job"})
		return
	}

	// Update film status to TRANSCODING
	tx, _ := h.queries.db.BeginTx(ctx, nil)
	h.queries.UpdateFilmStatus(ctx, tx, filmID, models.StatusTranscoding)
	tx.Commit()

	// Cache status in Redis
	h.redis.SetFilmStatus(ctx, filmID, models.StatusTranscoding)

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload confirmed. Transcoding started.",
		"job_id":  job.ID,
	})
}

// PublishFilm publishes a film (makes it publicly visible)
func (h *FilmHandler) PublishFilm(c *gin.Context) {
	idParam := c.Param("id")
	filmID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid film ID"})
		return
	}

	ctx := c.Request.Context()

	// Get film to verify ownership and status
	film, err := h.queries.GetFilmByID(ctx, filmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "film not found"})
		return
	}

	// Check ownership
	userID, _ := GetUserID(c)
	if film.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	// Can only publish READY films
	if film.Status != models.StatusReady {
		c.JSON(http.StatusBadRequest, gin.H{"error": "film must be in READY status to publish"})
		return
	}

	// Publish film
	tx, _ := h.queries.db.BeginTx(ctx, nil)
	if err := h.queries.PublishFilm(ctx, tx, filmID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish film"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Film published successfully",
	})
}

// GetPlaybackURL returns the HLS playback URL for a film
func (h *FilmHandler) GetPlaybackURL(c *gin.Context) {
	idParam := c.Param("id")
	filmID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid film ID"})
		return
	}

	ctx := c.Request.Context()

	// Get film
	film, err := h.queries.GetFilmByID(ctx, filmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "film not found"})
		return
	}

	// Check if film is ready
	if film.Status != models.StatusReady {
		c.JSON(http.StatusBadRequest, gin.H{"error": "film is not ready for playback"})
		return
	}

	// Increment view count asynchronously
	go h.queries.IncrementViewCount(ctx, filmID)

	// Get video assets
	assets, err := h.queries.GetVideoAssetsByFilmID(ctx, filmID)
	if err != nil {
		assets = []models.VideoAsset{}
	}

	// Return playback info
	c.JSON(http.StatusOK, gin.H{
		"hls_master_url": film.HLSMasterURL,
		"thumbnail_url":   film.ThumbnailURL,
		"assets":         assets,
	})
}
