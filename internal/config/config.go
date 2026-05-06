package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	GRPCPort    string
	DatabaseURL string

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool
}

func Load() (Config, error) {
	useSSL := false
	if raw := os.Getenv("S3_USE_SSL"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, errors.New("S3_USE_SSL must be true or false")
		}
		useSSL = parsed
	}

	databaseURL, err := loadDatabaseURL()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		GRPCPort:    getEnvOrDefault("GRPC_PORT", "50051"),
		DatabaseURL: databaseURL,
		S3Endpoint:  getEnvOrDefault("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey: getEnvOrDefault("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnvOrDefault("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:    getEnvOrDefault("S3_BUCKET", "fotos"),
		S3UseSSL:    useSSL,
	}

	if cfg.S3Endpoint == "" {
		return Config{}, errors.New("S3_ENDPOINT is required")
	}
	if cfg.S3AccessKey == "" {
		return Config{}, errors.New("S3_ACCESS_KEY is required")
	}
	if cfg.S3SecretKey == "" {
		return Config{}, errors.New("S3_SECRET_KEY is required")
	}
	if cfg.S3Bucket == "" {
		return Config{}, errors.New("S3_BUCKET is required")
	}

	return cfg, nil
}

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5433/user_service?sslmode=disable"

func loadDatabaseURL() (string, error) {
	raw, set := os.LookupEnv("DATABASE_URL")
	if !set {
		return defaultDatabaseURL, nil
	}
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("DATABASE_URL is required")
	}
	return strings.TrimSpace(raw), nil
}

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
