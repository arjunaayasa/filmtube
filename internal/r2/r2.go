package r2

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Storage paths in R2
const (
	OriginalPath = "original"
	ThumbnailPath = "thumb"
	HLSPath      = "hls"
)

type Client struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
	publicURL  string
}

// New creates a new Cloudflare R2 client (S3-compatible)
func New(endpoint, accessKey, secretKey, bucket, region, publicURL string) (*Client, error) {
	// Create custom resolver for R2 endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
			SigningRegion:     region,
		}, nil
	})

	// Load AWS config with custom endpoint
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			}, nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &Client{
		client:     client,
		uploader:   manager.NewUploader(client),
		downloader: manager.NewDownloader(client),
		bucket:     bucket,
		publicURL:  publicURL,
	}, nil
}

// ========== UPLOAD URL GENERATION ==========

// GeneratePresignedUploadURL creates a pre-signed URL for direct upload to R2
// The file will be uploaded to: original/{filmId}/source.mp4
func (c *Client) GeneratePresignedUploadURL(ctx context.Context, filmID uuid.UUID, expiration time.Duration) (string, error) {
	key := fmt.Sprintf("%s/%s/source.mp4", OriginalPath, filmID)

	presignClient := s3.NewPresignClient(c.client)

	presignedResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to presign put object: %w", err)
	}

	return presignedResult.URL, nil
}

// GeneratePresignedUploadURLForThumbnail creates a pre-signed URL for thumbnail upload
func (c *Client) GeneratePresignedUploadURLForThumbnail(ctx context.Context, filmID uuid.UUID, expiration time.Duration) (string, error) {
	key := fmt.Sprintf("%s/%s/poster.jpg", ThumbnailPath, filmID)

	presignClient := s3.NewPresignClient(c.client)

	presignedResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to presign put object: %w", err)
	}

	return presignedResult.URL, nil
}

// ========== FILE OPERATIONS ==========

// UploadFile uploads a file to R2
func (c *Client) UploadFile(ctx context.Context, key string, reader io.Reader, contentType string) error {
	_, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:          aws.String(key),
		Body:         reader,
		ContentType:  aws.String(contentType),
	})
	return err
}

// UploadHLSFile uploads an HLS file to R2
func (c *Client) UploadHLSFile(ctx context.Context, filmID uuid.UUID, quality, filename string, reader io.Reader) error {
	key := fmt.Sprintf("%s/%s/%s/%s", HLSPath, filmID, quality, filename)
	contentType := "application/x-mpegURL"
	if len(filename) > 4 && filename[len(filename)-3:] == ".ts" {
		contentType = "video/mp2t"
	}
	return c.UploadFile(ctx, key, reader, contentType)
}

// DownloadFile downloads a file from R2
func (c *Client) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	buffer := manager.NewWriteAtBuffer([]byte{})

	_, err := c.downloader.Download(ctx, buffer, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// DownloadOriginalVideo downloads the original video for transcoding
func (c *Client) DownloadOriginalVideo(ctx context.Context, filmID uuid.UUID) ([]byte, error) {
	key := fmt.Sprintf("%s/%s/source.mp4", OriginalPath, filmID)
	return c.DownloadFile(ctx, key)
}

// DeleteFilm removes all files associated with a film
func (c *Client) DeleteFilm(ctx context.Context, filmID uuid.UUID) error {
	// List all objects with the film ID prefix
	prefix := fmt.Sprintf("%s/%s/", OriginalPath, filmID)
	// Also need to clean up HLS and thumbnail paths
	paths := []string{
		fmt.Sprintf("%s/%s/", OriginalPath, filmID),
		fmt.Sprintf("%s/%s/", ThumbnailPath, filmID),
		fmt.Sprintf("%s/%s/", HLSPath, filmID),
	}

	for _, prefix := range paths {
		listOutput, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(c.bucket),
			Prefix: aws.String(prefix),
		})
		if err != nil {
			return err
		}

		for _, obj := range listOutput.Contents {
			_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(c.bucket),
				Key:    obj.Key,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ========== PUBLIC URL GENERATION ==========

// GetPublicURL returns the public URL for a file in R2
func (c *Client) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", c.publicURL, key)
}

// GetHLSMasterURL returns the public HLS master playlist URL for a film
func (c *Client) GetHLSMasterURL(filmID uuid.UUID) string {
	key := fmt.Sprintf("%s/%s/master.m3u8", HLSPath, filmID)
	return c.GetPublicURL(key)
}

// GetThumbnailURL returns the public thumbnail URL for a film
func (c *Client) GetThumbnailURL(filmID uuid.UUID) string {
	key := fmt.Sprintf("%s/%s/poster.jpg", ThumbnailPath, filmID)
	return c.GetPublicURL(key)
}

// GetOriginalVideoURL returns the public URL for original video (if accessible)
func (c *Client) GetOriginalVideoURL(filmID uuid.UUID) string {
	key := fmt.Sprintf("%s/%s/source.mp4", OriginalPath, filmID)
	return c.GetPublicURL(key)
}
