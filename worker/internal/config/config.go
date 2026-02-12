package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DatabaseURL string

	// Redis
	RedisURL     string
	RedisPassword string
	RedisDB       int

	// R2 (Cloudflare S3-compatible)
	R2Endpoint        string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2Region          string
	R2PublicURL       string

	// FFmpeg
	FFmpegPath string
	TempDir    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://filmtube:filmtube@localhost:5432/filmtube?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		R2Endpoint:        getEnv("R2_ENDPOINT", "https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com"),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:          getEnv("R2_BUCKET", "filmtube"),
		R2Region:          getEnv("R2_REGION", "auto"),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", "https://YOUR_R2_PUBLIC_DOMAIN"),
		FFmpegPath:         getEnv("FFMPEG_PATH", "ffmpeg"),
		TempDir:           getEnv("TEMP_DIR", os.TempDir()),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
