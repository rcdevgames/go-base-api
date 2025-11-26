package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config aggregates all runtime configuration consumed by the application.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Security SecurityConfig
	Storage  StorageConfig
}

// AppConfig holds HTTP server specific settings.
type AppConfig struct {
	Env               string
	Port              string
	LogDir            string
	RateLimitRequests int
	RateLimitWindow   time.Duration
	CORS              CORSConfig
}

// CORSConfig defines cross-origin resource sharing behaviour.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	ExposedHeaders   []string
	MaxAge           time.Duration
}

// DatabaseConfig describes the PostgreSQL connection details.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// SecurityConfig represents authentication and cryptography related settings.
type SecurityConfig struct {
	JWTSecret         string
	InterServiceToken string
	BcryptCost        int
	JWTExpiration     time.Duration
	RefreshSecret     string
	RefreshExpiration time.Duration
}

// StorageConfig defines file storage behavior.
type StorageConfig struct {
	Provider    string
	LocalDir    string
	CDNBaseURL  string
	S3Endpoint  string
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3UseSSL    bool
	S3BaseURL   string
}

// Load reads configuration from environment variables, returning an error when
// mandatory values are missing.
func Load() (*Config, error) {
	appPort, err := requireEnv("APP_PORT")
	if err != nil {
		return nil, err
	}

	user, err := requireEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	password, err := requireEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	name, err := requireEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	jwtSecret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	interToken, err := requireEnv("INTER_SERVICE_TOKEN")
	if err != nil {
		return nil, err
	}

	refreshSecret, err := requireEnv("REFRESH_SECRET")
	if err != nil {
		return nil, err
	}

	bcryptCost := getEnvAsInt("BCRYPT_COST", 10)

	return &Config{
		App: AppConfig{
			Env:               strings.ToLower(getEnvOrDefault("APP_ENV", "development")),
			Port:              appPort,
			LogDir:            getEnvOrDefault("LOG_DIR", "logs"),
			RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute),
			CORS: CORSConfig{
				AllowedOrigins:   getEnvAsCSV("CORS_ALLOWED_ORIGINS", []string{"*"}),
				AllowedMethods:   getEnvAsCSV("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
				AllowedHeaders:   getEnvAsCSV("CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-Internal-Token"}),
				AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
				ExposedHeaders:   getEnvAsCSV("CORS_EXPOSED_HEADERS", []string{}),
				MaxAge:           getEnvAsDuration("CORS_MAX_AGE", 24*time.Hour),
			},
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     user,
			Password: password,
			Name:     name,
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		Security: SecurityConfig{
			JWTSecret:         jwtSecret,
			InterServiceToken: interToken,
			BcryptCost:        bcryptCost,
			JWTExpiration:     time.Duration(getEnvAsInt("JWT_EXPIRATION_MINUTES", 60)) * time.Minute,
			RefreshSecret:     refreshSecret,
			RefreshExpiration: time.Duration(getEnvAsInt("REFRESH_EXPIRATION_MINUTES", 1440)) * time.Minute,
		},
		Storage: loadStorageConfig(),
	}, nil
}

// LoadStorageConfig exposes storage settings without requiring full config load.
func LoadStorageConfig() StorageConfig {
	return loadStorageConfig()
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s is required", key)
	}
	return value, nil
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return dur
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}

func getEnvAsCSV(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	var cleaned []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	if len(cleaned) == 0 {
		return fallback
	}
	return cleaned
}

func loadStorageConfig() StorageConfig {
	provider := strings.ToLower(getEnvOrDefault("STORAGE_PROVIDER", "local"))
	return StorageConfig{
		Provider:    provider,
		LocalDir:    getEnvOrDefault("LOCAL_UPLOAD_DIR", "upload"),
		CDNBaseURL:  getEnvOrDefault("CDN_BASE_URL", "http://localhost:8080/cdn"),
		S3Endpoint:  os.Getenv("S3_ENDPOINT"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3Region:    os.Getenv("S3_REGION"),
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
		S3UseSSL:    getEnvAsBool("S3_USE_SSL", true),
		S3BaseURL:   os.Getenv("S3_BASE_URL"),
	}
}
