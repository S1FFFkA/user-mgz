package testutil

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresEnv struct {
	Pool     *pgxpool.Pool
	ConnStr  string
	Teardown func(context.Context) error
}

func StartPostgres(ctx context.Context, migrationsDir string) (PostgresEnv, error) {
	connStr := getenv("DATABASE_URL")
	if connStr == "" {
		return PostgresEnv{}, fmt.Errorf("DATABASE_URL is required for integration tests (run `make compose-up` or set env)")
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return PostgresEnv{}, err
	}

	if err := waitForDB(ctx, pool); err != nil {
		pool.Close()
		return PostgresEnv{}, err
	}

	if migrationsDir != "" {
		abs, err := filepath.Abs(migrationsDir)
		if err != nil {
			pool.Close()
			return PostgresEnv{}, err
		}
		if err := applyMigrations(pool, abs); err != nil {
			pool.Close()
			return PostgresEnv{}, err
		}
	}

	return PostgresEnv{
		Pool:    pool,
		ConnStr: connStr,
		Teardown: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	}, nil
}

func waitForDB(ctx context.Context, pool *pgxpool.Pool) error {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if err := pool.Ping(ctx); err == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("db ping timeout")
}

func databaseURLForMigrate(connStr string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", fmt.Errorf("parse connection string: %w", err)
	}
	switch u.Scheme {
	case "postgres", "postgresql":
		u.Scheme = "pgx5"
		return u.String(), nil
	case "pgx5":
		return connStr, nil
	default:
		return "", fmt.Errorf("unsupported URL scheme %q for golang-migrate (need postgres:// or postgresql://)", u.Scheme)
	}
}

func applyMigrations(pool *pgxpool.Pool, migrationsDir string) error {
	src, err := iofs.New(os.DirFS(migrationsDir), ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	dbURL, err := databaseURLForMigrate(pool.Config().ConnString())
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func getenv(k string) string {
	v, _ := os.LookupEnv(k)
	return v
}
