package uploader

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rcdevgames/modular-monolith-clean/internal/storage"
)

var (
	// DefaultAllowedMIMEs enumerates safe defaults for uploads.
	DefaultAllowedMIMEs = []string{
		"image/png",
		"image/jpeg",
		"image/webp",
		"application/pdf",
	}

	extensionByMIME = map[string]string{
		"image/png":       "png",
		"image/jpeg":      "jpg",
		"image/webp":      "webp",
		"application/pdf": "pdf",
	}
)

// Options controls validation for uploads.
type Options struct {
	AllowedMIMETypes []string
	MaxSizeBytes     int64
	Prefix           string
}

// UploadBase64 validates a base64 payload and stores it using the configured storage backend.
func UploadBase64(ctx context.Context, backend storage.Storage, encoded string, opts Options) (string, error) {
	if backend == nil {
		return "", errors.New("missing storage backend")
	}
	data, contentType, err := decodeBase64(encoded)
	if err != nil {
		return "", err
	}
	if opts.MaxSizeBytes > 0 && int64(len(data)) > opts.MaxSizeBytes {
		return "", fmt.Errorf("file exceeds maximum size of %d bytes", opts.MaxSizeBytes)
	}
	allowed := opts.AllowedMIMETypes
	if len(allowed) == 0 {
		allowed = DefaultAllowedMIMEs
	}
	if !contains(allowed, contentType) {
		return "", fmt.Errorf("mime type %s not allowed", contentType)
	}
	objectName := buildObjectName(opts.Prefix, contentType)
	return backend.Upload(ctx, objectName, data, contentType)
}

func decodeBase64(encoded string) ([]byte, string, error) {
	if encoded == "" {
		return nil, "", errors.New("empty payload")
	}
	parts := strings.SplitN(encoded, ",", 2)
	if len(parts) == 2 && strings.Contains(parts[0], "base64") {
		encoded = parts[1]
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", fmt.Errorf("decode base64: %w", err)
	}
	contentType := http.DetectContentType(decoded)
	return decoded, contentType, nil
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}

func buildObjectName(prefix, mime string) string {
	ext := extensionByMIME[mime]
	if ext == "" {
		ext = "bin"
	}
	id := uuid.New().String()
	timestamp := time.Now().UTC().Format("20060102")
	key := fmt.Sprintf("%s.%s", id, ext)
	if prefix != "" {
		return path.Join(prefix, timestamp, key)
	}
	return path.Join(timestamp, key)
}
