//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/integration/testutil"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userrepo "github.com/S1FFFkA/user-mgz/internal/repository/user"
	"github.com/google/uuid"
)

func TestUserRepository_CreateGetUpdateDelete(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	r := userrepo.New(env.Pool)

	created, err := r.CreateUser(ctx, domain.User{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		ToilerScore:           7,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoURL:       "u",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatalf("expected id")
	}

	got, err := r.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Email != "a@b.com" {
		t.Fatalf("unexpected email: %s", got.Email)
	}

	updated, err := r.UpdateUser(ctx, domain.User{
		ID:                    created.ID,
		FirstName:             "AA",
		LastName:              "BB",
		Email:                 "aa@bb.com",
		BirthDate:             got.BirthDate,
		ToilerScore:           8,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "k2",
		PrimaryPhotoURL:       "u2",
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.Email != "aa@bb.com" {
		t.Fatalf("unexpected updated email: %s", updated.Email)
	}

	if err := r.DeleteUser(ctx, created.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	_, err = r.GetUserByID(ctx, created.ID)
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeNotFound) {
		t.Fatalf("expected not found, got: %v", err)
	}
}

func TestUserRepository_ListUsers_WithCityFilter(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	// insert city
	_, err = env.Pool.Exec(ctx, `INSERT INTO cities (name) VALUES ('Moscow')`)
	if err != nil {
		t.Fatalf("insert city: %v", err)
	}
	var cityID int64
	if err := env.Pool.QueryRow(ctx, `SELECT id FROM cities WHERE name='Moscow' LIMIT 1`).Scan(&cityID); err != nil {
		t.Fatalf("select city id: %v", err)
	}

	r := userrepo.New(env.Pool)
	for i := 0; i < 3; i++ {
		u := domain.User{
			FirstName:             "A",
			LastName:              "B",
			Email:                 uuid.Must(uuid.NewV7()).String() + "@b.com",
			BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			ToilerScore:           7,
			Sex:                   domain.SexMale,
			PrimaryPhotoObjectKey: "k",
			PrimaryPhotoURL:       "u",
		}
		if i%2 == 0 {
			u.CityID = &cityID
		}
		if _, err := r.CreateUser(ctx, u); err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
	}

	all, err := r.ListUsers(ctx, 20, 0, nil)
	if err != nil {
		t.Fatalf("ListUsers all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 users, got %d", len(all))
	}

	filtered, err := r.ListUsers(ctx, 20, 0, &cityID)
	if err != nil {
		t.Fatalf("ListUsers city: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 users for city, got %d", len(filtered))
	}
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	r := userrepo.New(env.Pool)
	_, err = r.GetUserByID(ctx, uuid.Must(uuid.NewV7()))
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeNotFound) {
		t.Fatalf("expected not found, got: %v", err)
	}
}

func TestUserRepository_DeleteUser_NotFound(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	r := userrepo.New(env.Pool)
	err = r.DeleteUser(ctx, uuid.Must(uuid.NewV7()))
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeNotFound) {
		t.Fatalf("expected not found, got: %v", err)
	}
}
