//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/integration/testutil"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userrepo "github.com/S1FFFkA/user-mgz/internal/repository/user"
	photorepo "github.com/S1FFFkA/user-mgz/internal/repository/userphoto"
	"github.com/google/uuid"
)

func createUser(t *testing.T, ctx context.Context, ur *userrepo.UserRepository, email string) uuid.UUID {
	t.Helper()
	u, err := ur.CreateUser(ctx, domain.User{
		FirstName:             "A",
		LastName:              "B",
		Email:                 email,
		BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		ToilerScore:           7,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoURL:       "u",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return u.ID
}

func TestUserPhotoRepository_InsertListGetDelete(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	ur := userrepo.New(env.Pool)
	pr := photorepo.New(env.Pool)

	uid := createUser(t, ctx, ur, "a@b.com")

	if err := pr.InsertExtraPhoto(ctx, uid, domain.UserPhoto{ObjectKey: "k1", URL: "u1", Position: 1}); err != nil {
		t.Fatalf("InsertExtraPhoto: %v", err)
	}
	if err := pr.InsertExtraPhoto(ctx, uid, domain.UserPhoto{ObjectKey: "k2", URL: "u2", Position: 2}); err != nil {
		t.Fatalf("InsertExtraPhoto: %v", err)
	}

	list, err := pr.ListExtraPhotos(ctx, uid)
	if err != nil {
		t.Fatalf("ListExtraPhotos: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 photos, got %d", len(list))
	}

	got, err := pr.GetExtraPhotoByID(ctx, uid, list[0].ID)
	if err != nil {
		t.Fatalf("GetExtraPhotoByID: %v", err)
	}
	if got.UserID != uid {
		t.Fatalf("unexpected user id")
	}

	if err := pr.DeleteExtraPhoto(ctx, uid, list[0].ID); err != nil {
		t.Fatalf("DeleteExtraPhoto: %v", err)
	}
	_, err = pr.GetExtraPhotoByID(ctx, uid, list[0].ID)
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeNotFound) {
		t.Fatalf("expected not found, got: %v", err)
	}
}

func TestUserPhotoRepository_DeleteExtraByPosition_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	ur := userrepo.New(env.Pool)
	pr := photorepo.New(env.Pool)
	uid := createUser(t, ctx, ur, "x@y.com")

	if err := pr.DeleteExtraPhotoByPosition(ctx, uid, 1); err != nil {
		t.Fatalf("DeleteExtraPhotoByPosition: %v", err)
	}
}

func TestUserPhotoRepository_SetPrimaryPhoto_NotFoundUser(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	pr := photorepo.New(env.Pool)
	uid := uuid.Must(uuid.NewV7())
	err = pr.SetPrimaryPhoto(ctx, uid, "k", "u")
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeNotFound) {
		t.Fatalf("expected not found, got: %v", err)
	}
}

func TestUserPhotoRepository_DeleteAllExtraPhotos(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	ur := userrepo.New(env.Pool)
	pr := photorepo.New(env.Pool)
	uid := createUser(t, ctx, ur, "z@z.com")

	_ = pr.InsertExtraPhoto(ctx, uid, domain.UserPhoto{ObjectKey: "k1", URL: "u1", Position: 1})
	_ = pr.InsertExtraPhoto(ctx, uid, domain.UserPhoto{ObjectKey: "k2", URL: "u2", Position: 2})

	if err := pr.DeleteAllExtraPhotos(ctx, uid); err != nil {
		t.Fatalf("DeleteAllExtraPhotos: %v", err)
	}
	list, err := pr.ListExtraPhotos(ctx, uid)
	if err != nil {
		t.Fatalf("ListExtraPhotos: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 photos, got %d", len(list))
	}
}

func TestUserPhotoRepository_DeleteExtraPhoto_NotFound(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	ur := userrepo.New(env.Pool)
	pr := photorepo.New(env.Pool)
	uid := createUser(t, ctx, ur, "nf@nf.com")

	err = pr.DeleteExtraPhoto(ctx, uid, 999999)
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeNotFound) {
		t.Fatalf("expected not found, got: %v", err)
	}
}

func TestUserPhotoRepository_GetExtraPhotoByID_NotFound(t *testing.T) {
	ctx := context.Background()
	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	ur := userrepo.New(env.Pool)
	pr := photorepo.New(env.Pool)
	uid := createUser(t, ctx, ur, "nf2@nf2.com")

	_, err = pr.GetExtraPhotoByID(ctx, uid, 12345)
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeNotFound) {
		t.Fatalf("expected not found, got: %v", err)
	}
}
