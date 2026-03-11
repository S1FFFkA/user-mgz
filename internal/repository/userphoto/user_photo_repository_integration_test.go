//go:build integration

package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	userphotorepo "github.com/S1FFFkA/user-mgz/internal/repository/userphoto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPhotoDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pg, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("user_service"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = pg.Terminate(ctx)
	})

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("container dsn: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool new: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	if err = applyPhotoMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return pool
}

func applyPhotoMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return os.ErrNotExist
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	sqlPath := filepath.Join(root, "migrations", "001_init.up.sql")
	data, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(data))
	return err
}

func insertTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	_, err := pool.Exec(ctx, `
INSERT INTO users (
  id, first_name, last_name, email, birth_date, toiler_score, sex, primary_photo_object_key, primary_photo_url
) VALUES ($1, 'Ivan', 'Petrov', $2, '2000-01-01', 7, 'male', 'users/u/primary/old.jpg', 'users/u/primary/old.jpg')
`, id, "user-"+id.String()+"@example.com")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func TestUserPhotoRepositoryFlow(t *testing.T) {
	ctx := context.Background()
	pool := setupPhotoDB(t)
	repo := userphotorepo.NewRepository(pool)
	userID := insertTestUser(t, ctx, pool)

	if err := repo.SetPrimaryPhoto(ctx, userID, "users/u/primary/new.jpg", "users/u/primary/new.jpg"); err != nil {
		t.Fatalf("SetPrimaryPhoto: %v", err)
	}

	if err := repo.UpsertExtraPhotoByPosition(ctx, userID, domain.UserPhoto{
		ObjectKey: "users/u/extra/1/a.jpg",
		URL:       "users/u/extra/1/a.jpg",
		Position:  1,
	}); err != nil {
		t.Fatalf("UpsertExtraPhotoByPosition insert: %v", err)
	}
	if err := repo.UpsertExtraPhotoByPosition(ctx, userID, domain.UserPhoto{
		ObjectKey: "users/u/extra/1/b.jpg",
		URL:       "users/u/extra/1/b.jpg",
		Position:  1,
	}); err != nil {
		t.Fatalf("UpsertExtraPhotoByPosition update: %v", err)
	}

	list, err := repo.ListExtraPhotos(ctx, userID)
	if err != nil {
		t.Fatalf("ListExtraPhotos: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 photo, got %d", len(list))
	}

	photoID := list[0].ID
	if err = repo.ReplaceExtraPhoto(ctx, userID, photoID, domain.UserPhoto{
		ObjectKey: "users/u/extra/1/c.jpg",
		URL:       "users/u/extra/1/c.jpg",
		Position:  1,
	}); err != nil {
		t.Fatalf("ReplaceExtraPhoto: %v", err)
	}

	if err = repo.ReplaceExtraPhotos(ctx, userID, []domain.UserPhoto{
		{ObjectKey: "users/u/extra/1/d.jpg", URL: "users/u/extra/1/d.jpg", Position: 1},
		{ObjectKey: "users/u/extra/2/e.jpg", URL: "users/u/extra/2/e.jpg", Position: 2},
	}); err != nil {
		t.Fatalf("ReplaceExtraPhotos: %v", err)
	}

	list, err = repo.ListExtraPhotos(ctx, userID)
	if err != nil {
		t.Fatalf("ListExtraPhotos after replace: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 photos, got %d", len(list))
	}

	one, err := repo.GetExtraPhotoByID(ctx, userID, list[0].ID)
	if err != nil {
		t.Fatalf("GetExtraPhotoByID: %v", err)
	}
	if one.CreatedAt.Equal(time.Time{}) {
		t.Fatalf("expected created_at to be set")
	}
}
