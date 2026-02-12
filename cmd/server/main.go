package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arjunaayasa/filmtube/internal/api"
	"github.com/arjunaayasa/filmtube/internal/auth"
	"github.com/arjunaayasa/filmtube/internal/config"
	"github.com/arjunaayasa/filmtube/internal/db"
	"github.com/arjunaayasa/filmtube/internal/models"
	"github.com/arjunaayasa/filmtube/internal/r2"
	"github.com/arjunaayasa/filmtube/internal/redis"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func main() {
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

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connected successfully")

	// Initialize Redis
	redisClient, err := redis.New(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected successfully")

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
	log.Println("R2 client initialized successfully")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)

	// Initialize queries
	queries := db.NewQueries(database)

	// Initialize handlers
	authHandler := api.NewAuthHandler(queries, jwtManager)
	filmHandler := api.NewFilmHandler(queries, r2Client, redisClient, int(cfg.UploadURLExpiration.Minutes()))

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:          86400,
	})
	router.Use(func(c *gin.Context) {
		corsHandler.HandlerFunc(c.Writer, c.Request)
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "filmtube-api",
		})
	})

	// Public routes
	public := router.Group("/api")
	{
		// Auth routes
		auth := public.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Public film routes (browse)
		films := public.Group("/films")
		{
			films.GET("", filmHandler.ListFilms)
			films.GET("/:id", filmHandler.GetFilm)
			films.GET("/:id/playback", filmHandler.GetPlaybackURL)
		}
	}

	// Protected routes (require authentication)
	protected := router.Group("/api")
	protected.Use(api.AuthMiddleware(jwtManager))
	{
		// User routes
		protected.GET("/auth/me", authHandler.GetMe)

		// Film management routes (require creator role)
		films := protected.Group("/films")
		films.Use(api.RequireCreator())
		{
			films.POST("", filmHandler.CreateFilm)
			films.POST("/:id/upload-url", filmHandler.GetUploadURL)
			films.POST("/:id/confirm-upload", filmHandler.ConfirmUpload)
			films.POST("/:id/publish", filmHandler.PublishFilm)
		}
	}

	// Start server
	addr := ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}
