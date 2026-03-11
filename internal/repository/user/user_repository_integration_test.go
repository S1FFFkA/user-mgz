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
	userrepo "github.com/S1FFFkA/user-mgz/internal/repository/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupDB(t *testing.T) *pgxpool.Pool {
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

	if err = applyMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return pool
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
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

func TestUserRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	pool := setupDB(t)
	repo := userrepo.NewRepository(pool)

	created, err := repo.CreateUser(ctx, domain.User{
		FirstName:             "Ivan",
		LastName:              "Petrov",
		Email:                 "ivan.petrov@example.com",
		BirthDate:             time.Date(1998, 10, 10, 0, 0, 0, 0, time.UTC),
		ToilerScore:           8,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "users/x/primary/p.jpg",
		PrimaryPhotoURL:       "users/x/primary/p.jpg",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatalf("expected non-nil user id")
	}

	got, err := repo.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Email != "ivan.petrov@example.com" {
		t.Fatalf("unexpected email: %s", got.Email)
	}

	got.LastName = "Sidorov"
	updated, err := repo.UpdateUser(ctx, got)
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.LastName != "Sidorov" {
		t.Fatalf("expected updated last name, got %s", updated.LastName)
	}

	users, err := repo.ListUsers(ctx, 20, 0, nil)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("expected users list to be non-empty")
	}

	if err = repo.DeleteUser(ctx, created.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if _, err = repo.GetUserByID(ctx, created.ID); err == nil {
		t.Fatalf("expected error after delete")
	}
}
