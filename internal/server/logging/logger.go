package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// Setup configures the default logger to write to stdout and a rotating file inside logDir.
// Caller should invoke the returned cleanup function when shutting down.
func Setup(logDir string) (func(), error) {
	if logDir == "" {
		logDir = "logs"
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	pattern := filepath.Join(logDir, "app-%Y-%m-%d.log")
	writer, err := rotatelogs.New(
		pattern,
		rotatelogs.WithLinkName(filepath.Join(logDir, "current.log")),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(30*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	log.SetOutput(io.MultiWriter(os.Stdout, writer))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return func() {
		_ = writer.Close()
	}, nil
}
