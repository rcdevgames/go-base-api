package logging

import (
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global structured logger instance.
var Logger *zap.Logger

// Setup configures the structured logger with console and rotating file outputs.
// Returns a cleanup function to be called on shutdown.
func Setup(logDir string) (func(), error) {
	if logDir == "" {
		logDir = "logs"
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	// Create rotating file writer
	pattern := filepath.Join(logDir, "app-%Y-%m-%d.log")
	fileWriter, err := rotatelogs.New(
		pattern,
		rotatelogs.WithLinkName(filepath.Join(logDir, "current.log")),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(30*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	// Console encoder for development
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	// JSON encoder for production
	jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// File core with JSON encoding
	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(fileWriter), zapcore.InfoLevel)

	// Console core with console encoding
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)

	// Combine cores
	core := zapcore.NewTee(fileCore, consoleCore)

	// Create logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return func() {
		_ = Logger.Sync()
		_ = fileWriter.Close()
	}, nil
}

// GetLogger returns the global logger instance.
func GetLogger() *zap.Logger {
	return Logger
}
