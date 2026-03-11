package config

import (
	"errors"
	"os"
	"strconv"
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

	cfg := Config{
		GRPCPort:    getEnvOrDefault("GRPC_PORT", "50051"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		S3Endpoint:  os.Getenv("S3_ENDPOINT"),
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3UseSSL:    useSSL,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
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

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
