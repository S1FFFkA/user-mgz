package user

import (
	"context"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/S1FFFkA/user-mgz/test/testutil"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type mockUserRepo struct {
	createFn func(ctx context.Context, user domain.User) (domain.User, error)
	getFn    func(ctx context.Context, userID uuid.UUID) (domain.User, error)
	updateFn func(ctx context.Context, user domain.User) (domain.User, error)
	deleteFn func(ctx context.Context, userID uuid.UUID) error
	listFn   func(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.createFn(ctx, user)
}
func (m *mockUserRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	return m.getFn(ctx, userID)
}
func (m *mockUserRepo) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.updateFn(ctx, user)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return m.deleteFn(ctx, userID)
}
func (m *mockUserRepo) ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error) {
	return m.listFn(ctx, limit, offset, cityID)
}

type mockUserPhotoRepo struct {
	setPrimaryFn       func(ctx context.Context, userID uuid.UUID, objectKey, url string) error
	deleteAllFn        func(ctx context.Context, userID uuid.UUID) error
	deletePhotoFn      func(ctx context.Context, userID uuid.UUID, photoID int64) error
	deleteByPositionFn func(ctx context.Context, userID uuid.UUID, position int16) error
	insertFn           func(ctx context.Context, userID uuid.UUID, photo domain.UserPhoto) error
	getByIDFn          func(ctx context.Context, userID uuid.UUID, photoID int64) (domain.UserPhoto, error)
	listFn             func(ctx context.Context, userID uuid.UUID) ([]domain.UserPhoto, error)
}

func (m *mockUserPhotoRepo) SetPrimaryPhoto(ctx context.Context, userID uuid.UUID, objectKey, url string) error {
	return m.setPrimaryFn(ctx, userID, objectKey, url)
}
func (m *mockUserPhotoRepo) DeleteAllExtraPhotos(ctx context.Context, userID uuid.UUID) error {
	return m.deleteAllFn(ctx, userID)
}
func (m *mockUserPhotoRepo) DeleteExtraPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error {
	return m.deletePhotoFn(ctx, userID, photoID)
}
func (m *mockUserPhotoRepo) DeleteExtraPhotoByPosition(ctx context.Context, userID uuid.UUID, position int16) error {
	return m.deleteByPositionFn(ctx, userID, position)
}
func (m *mockUserPhotoRepo) InsertExtraPhoto(ctx context.Context, userID uuid.UUID, photo domain.UserPhoto) error {
	return m.insertFn(ctx, userID, photo)
}
func (m *mockUserPhotoRepo) GetExtraPhotoByID(ctx context.Context, userID uuid.UUID, photoID int64) (domain.UserPhoto, error) {
	return m.getByIDFn(ctx, userID, photoID)
}
func (m *mockUserPhotoRepo) ListExtraPhotos(ctx context.Context, userID uuid.UUID) ([]domain.UserPhoto, error) {
	return m.listFn(ctx, userID)
}

func noopPhotoRepo() *mockUserPhotoRepo {
	return &mockUserPhotoRepo{
		setPrimaryFn:       func(context.Context, uuid.UUID, string, string) error { return nil },
		deleteAllFn:        func(context.Context, uuid.UUID) error { return nil },
		deletePhotoFn:      func(context.Context, uuid.UUID, int64) error { return nil },
		deleteByPositionFn: func(context.Context, uuid.UUID, int16) error { return nil },
		insertFn:           func(context.Context, uuid.UUID, domain.UserPhoto) error { return nil },
		getByIDFn:          func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
		listFn:             func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
	}
}

func TestCreateUserSuccess(t *testing.T) {
	ctx := context.Background()
	uid := uuid.Must(uuid.NewV7())
	now := time.Now()
	extras := []domain.UserPhoto{{ObjectKey: "k1", URL: "u1", Position: 1}}

	svc := NewService(
		&mockUserRepo{
			createFn: func(_ context.Context, user domain.User) (domain.User, error) {
				user.ID = uid
				user.CreatedAt = now
				user.UpdatedAt = now
				return user, nil
			},
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		&mockUserPhotoRepo{
			setPrimaryFn:       func(context.Context, uuid.UUID, string, string) error { return nil },
			deleteAllFn:        func(context.Context, uuid.UUID) error { return nil },
			deletePhotoFn:      func(context.Context, uuid.UUID, int64) error { return nil },
			deleteByPositionFn: func(context.Context, uuid.UUID, int16) error { return nil },
			insertFn: func(_ context.Context, gotUserID uuid.UUID, photo domain.UserPhoto) error {
				if gotUserID != uid || photo.Position != 1 {
					t.Fatalf("unexpected insert args")
				}
				return nil
			},
			getByIDFn: func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:    func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return extras, nil },
		},
		testutil.PassthroughTxManager{},
		zap.NewNop(),
	)

	user := domain.User{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		ToilerScore:           7,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "users/u/primary/a.jpg",
		PrimaryPhotoURL:       "users/u/primary/a.jpg",
		ExtraPhotos:           extras,
	}

	created, err := svc.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	if created.ID != uid {
		t.Fatalf("unexpected id: %s", created.ID)
	}
	if len(created.ExtraPhotos) != 1 {
		t.Fatalf("expected 1 extra photo, got %d", len(created.ExtraPhotos))
	}
}

func TestCreateUserPropagatesRepositoryError(t *testing.T) {
	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) {
				return domain.User{}, domain.InternalError(nil)
			},
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		noopPhotoRepo(),
		testutil.PassthroughTxManager{},
		zap.NewNop(),
	)

	_, err := svc.CreateUser(context.Background(), domain.User{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		ToilerScore:           8,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoURL:       "u",
	})
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeInternal) {
		t.Fatalf("expected internal error, got: %v", err)
	}
}

func TestGetUserInvalidID(t *testing.T) {
	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		noopPhotoRepo(),
		testutil.PassthroughTxManager{},
		zap.NewNop(),
	)
	_, err := svc.GetUser(context.Background(), uuid.Nil)
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument) {
		t.Fatalf("expected invalid argument, got: %v", err)
	}
}

func TestListUsersInvalidCityID(t *testing.T) {
	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		noopPhotoRepo(),
		testutil.PassthroughTxManager{},
		zap.NewNop(),
	)
	cityID := int64(-1)
	_, err := svc.ListUsers(context.Background(), 20, 0, &cityID)
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument) {
		t.Fatalf("expected invalid argument, got: %v", err)
	}
}
