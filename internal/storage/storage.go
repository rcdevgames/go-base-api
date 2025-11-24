package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/rcdevgames/modular-monolith-clean/internal/config"
)

// Storage describes capabilities for persisting binary objects and returning public URLs.
type Storage interface {
	Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error)
}

// New builds a storage backend based on configuration.
func New(cfg config.StorageConfig) (Storage, error) {
	switch strings.ToLower(cfg.Provider) {
	case "s3", "minio":
		return newS3Storage(cfg)
	case "local", "":
		fallthrough
	default:
		return newLocalStorage(cfg)
	}
}

// BuildObjectURL safely concatenates base URLs with object paths.
func BuildObjectURL(base, objectName string) string {
	base = strings.TrimRight(base, "/")
	if base == "" {
		return "/" + strings.TrimLeft(objectName, "/")
	}
	return fmt.Sprintf("%s/%s", base, strings.TrimLeft(objectName, "/"))
}

type localStorage struct {
	baseDir string
	baseURL string
}

func newLocalStorage(cfg config.StorageConfig) (Storage, error) {
	baseDir := cfg.LocalDir
	if baseDir == "" {
		baseDir = "upload"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &localStorage{baseDir: baseDir, baseURL: cfg.CDNBaseURL}, nil
}

func (s *localStorage) Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	path := filepath.Join(s.baseDir, filepath.FromSlash(objectName))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("ensure dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return BuildObjectURL(s.baseURL, objectName), nil
}

type s3Storage struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

func newS3Storage(cfg config.StorageConfig) (Storage, error) {
	if cfg.S3Endpoint == "" || cfg.S3Bucket == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" {
		return nil, fmt.Errorf("missing S3 configuration")
	}
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{Region: cfg.S3Region}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}
	return &s3Storage{client: client, bucket: cfg.S3Bucket, baseURL: cfg.S3BaseURL}, nil
}

func (s *s3Storage) Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	baseURL := s.baseURL
	if baseURL == "" {
		scheme := "http"
		if s.client.EndpointURL().Scheme == "https" {
			scheme = "https"
		}
		baseURL = fmt.Sprintf("%s://%s/%s", scheme, s.client.EndpointURL().Host, s.bucket)
	}
	return BuildObjectURL(baseURL, objectName), nil
}
