package config

import (
	"os"
	"testing"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("S3_ENDPOINT", "minio:9000")
	t.Setenv("S3_ACCESS_KEY", "a")
	t.Setenv("S3_SECRET_KEY", "b")
	t.Setenv("S3_BUCKET", "bucket")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoad_ParsesS3UseSSL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("S3_ENDPOINT", "minio:9000")
	t.Setenv("S3_ACCESS_KEY", "a")
	t.Setenv("S3_SECRET_KEY", "b")
	t.Setenv("S3_BUCKET", "bucket")
	t.Setenv("S3_USE_SSL", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.S3UseSSL != true {
		t.Fatalf("expected true")
	}

	// and invalid
	_ = os.Setenv("S3_USE_SSL", "nope")
	_, err = Load()
	if err == nil {
		t.Fatalf("expected error")
	}
}
