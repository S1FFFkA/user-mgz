package user

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
	setPrimaryFn func(ctx context.Context, userID uuid.UUID, objectKey, url string) error
	replaceAllFn func(ctx context.Context, userID uuid.UUID, photos []domain.UserPhoto) error
	replaceOneFn func(ctx context.Context, userID uuid.UUID, photoID int64, photo domain.UserPhoto) error
	upsertPosFn  func(ctx context.Context, userID uuid.UUID, photo domain.UserPhoto) error
	getByIDFn    func(ctx context.Context, userID uuid.UUID, photoID int64) (domain.UserPhoto, error)
	listFn       func(ctx context.Context, userID uuid.UUID) ([]domain.UserPhoto, error)
}

func (m *mockUserPhotoRepo) SetPrimaryPhoto(ctx context.Context, userID uuid.UUID, objectKey, url string) error {
	return m.setPrimaryFn(ctx, userID, objectKey, url)
}
func (m *mockUserPhotoRepo) ReplaceExtraPhotos(ctx context.Context, userID uuid.UUID, photos []domain.UserPhoto) error {
	return m.replaceAllFn(ctx, userID, photos)
}
func (m *mockUserPhotoRepo) ReplaceExtraPhoto(ctx context.Context, userID uuid.UUID, photoID int64, photo domain.UserPhoto) error {
	return m.replaceOneFn(ctx, userID, photoID, photo)
}
func (m *mockUserPhotoRepo) UpsertExtraPhotoByPosition(ctx context.Context, userID uuid.UUID, photo domain.UserPhoto) error {
	return m.upsertPosFn(ctx, userID, photo)
}
func (m *mockUserPhotoRepo) GetExtraPhotoByID(ctx context.Context, userID uuid.UUID, photoID int64) (domain.UserPhoto, error) {
	return m.getByIDFn(ctx, userID, photoID)
}
func (m *mockUserPhotoRepo) ListExtraPhotos(ctx context.Context, userID uuid.UUID) ([]domain.UserPhoto, error) {
	return m.listFn(ctx, userID)
}

type mockS3Repo struct {
	presignPutFn func(ctx context.Context, objectKey, contentType string, contentLength int64, expiresIn time.Duration) (string, error)
	presignGetFn func(ctx context.Context, objectKey, fileName string, asAttachment bool, expiresIn time.Duration) (string, error)
	deleteObjFn  func(ctx context.Context, objectKey string) error
}

func (m *mockS3Repo) PresignPutURL(ctx context.Context, objectKey, contentType string, contentLength int64, expiresIn time.Duration) (string, error) {
	return m.presignPutFn(ctx, objectKey, contentType, contentLength, expiresIn)
}
func (m *mockS3Repo) PresignGetURL(ctx context.Context, objectKey, fileName string, asAttachment bool, expiresIn time.Duration) (string, error) {
	return m.presignGetFn(ctx, objectKey, fileName, asAttachment, expiresIn)
}
func (m *mockS3Repo) DeleteObject(ctx context.Context, objectKey string) error {
	return m.deleteObjFn(ctx, objectKey)
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
			setPrimaryFn: func(context.Context, uuid.UUID, string, string) error { return nil },
			replaceAllFn: func(_ context.Context, gotUserID uuid.UUID, photos []domain.UserPhoto) error {
				if gotUserID != uid || len(photos) != 1 {
					t.Fatalf("unexpected replace args")
				}
				return nil
			},
			replaceOneFn: func(context.Context, uuid.UUID, int64, domain.UserPhoto) error { return nil },
			upsertPosFn:  func(context.Context, uuid.UUID, domain.UserPhoto) error { return nil },
			getByIDFn:    func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:       func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return extras, nil },
		},
		&mockS3Repo{
			presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
			presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
			deleteObjFn:  func(context.Context, string) error { return nil },
		},
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

func TestCreateUserInvalidScore(t *testing.T) {
	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		&mockUserPhotoRepo{
			setPrimaryFn: func(context.Context, uuid.UUID, string, string) error { return nil },
			replaceAllFn: func(context.Context, uuid.UUID, []domain.UserPhoto) error { return nil },
			replaceOneFn: func(context.Context, uuid.UUID, int64, domain.UserPhoto) error { return nil },
			upsertPosFn:  func(context.Context, uuid.UUID, domain.UserPhoto) error { return nil },
			getByIDFn:    func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:       func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
		},
		&mockS3Repo{
			presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
			presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
			deleteObjFn:  func(context.Context, string) error { return nil },
		},
	)

	_, err := svc.CreateUser(context.Background(), domain.User{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		ToilerScore:           11,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoURL:       "u",
	})
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument) {
		t.Fatalf("expected invalid argument, got: %v", err)
	}
}

func TestGetUserPhotoUploadURLSuccess(t *testing.T) {
	var gotObjectKey string
	userID := uuid.Must(uuid.NewV7())

	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		&mockUserPhotoRepo{
			setPrimaryFn: func(context.Context, uuid.UUID, string, string) error { return nil },
			replaceAllFn: func(context.Context, uuid.UUID, []domain.UserPhoto) error { return nil },
			replaceOneFn: func(context.Context, uuid.UUID, int64, domain.UserPhoto) error { return nil },
			upsertPosFn:  func(context.Context, uuid.UUID, domain.UserPhoto) error { return nil },
			getByIDFn:    func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:       func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
		},
		&mockS3Repo{
			presignPutFn: func(_ context.Context, objectKey, _ string, _ int64, _ time.Duration) (string, error) {
				gotObjectKey = objectKey
				return "https://upload", nil
			},
			presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
			deleteObjFn:  func(context.Context, string) error { return nil },
		},
	)

	pos := int16(2)
	ticket, err := svc.GetUserPhotoUploadURL(context.Background(), domain.UploadPhotoRequest{
		UserID:        userID,
		PhotoType:     domain.PhotoTypeExtra,
		ExtraPosition: &pos,
		ContentType:   "image/jpeg",
		ContentLength: 1024,
	})
	if err != nil {
		t.Fatalf("GetUserPhotoUploadURL error: %v", err)
	}
	if ticket.UploadURL != "https://upload" {
		t.Fatalf("unexpected upload url: %s", ticket.UploadURL)
	}
	if !strings.Contains(gotObjectKey, "/extra/2/") {
		t.Fatalf("unexpected object key: %s", gotObjectKey)
	}
}

func TestConfirmUserPhotoUploadExtraUsesUpsert(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	upsertCalled := false

	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			getFn: func(context.Context, uuid.UUID) (domain.User, error) {
				return domain.User{
					ID:                    userID,
					FirstName:             "A",
					LastName:              "B",
					Email:                 "a@b.com",
					BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
					ToilerScore:           7,
					Sex:                   domain.SexMale,
					PrimaryPhotoObjectKey: "k",
					PrimaryPhotoURL:       "u",
				}, nil
			},
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		&mockUserPhotoRepo{
			setPrimaryFn: func(context.Context, uuid.UUID, string, string) error { return nil },
			replaceAllFn: func(context.Context, uuid.UUID, []domain.UserPhoto) error { return nil },
			replaceOneFn: func(context.Context, uuid.UUID, int64, domain.UserPhoto) error { return nil },
			upsertPosFn: func(_ context.Context, gotUserID uuid.UUID, photo domain.UserPhoto) error {
				upsertCalled = true
				if gotUserID != userID || photo.Position != 1 {
					t.Fatalf("unexpected upsert args")
				}
				return nil
			},
			getByIDFn: func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:    func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
		},
		&mockS3Repo{
			presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
			presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
			deleteObjFn:  func(context.Context, string) error { return nil },
		},
	)

	pos := int16(1)
	_, err := svc.ConfirmUserPhotoUpload(context.Background(), domain.ConfirmPhotoUploadRequest{
		UserID:        userID,
		PhotoType:     domain.PhotoTypeExtra,
		ExtraPosition: &pos,
		ObjectKey:     "users/u/extra/1/new.jpg",
	})
	if err != nil {
		t.Fatalf("ConfirmUserPhotoUpload error: %v", err)
	}
	if !upsertCalled {
		t.Fatalf("expected upsert to be called")
	}
}

func TestDeleteUserPhotoNotAllowed(t *testing.T) {
	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		&mockUserPhotoRepo{
			setPrimaryFn: func(context.Context, uuid.UUID, string, string) error { return nil },
			replaceAllFn: func(context.Context, uuid.UUID, []domain.UserPhoto) error { return nil },
			replaceOneFn: func(context.Context, uuid.UUID, int64, domain.UserPhoto) error { return nil },
			upsertPosFn:  func(context.Context, uuid.UUID, domain.UserPhoto) error { return nil },
			getByIDFn:    func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:       func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
		},
		&mockS3Repo{
			presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
			presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
			deleteObjFn:  func(context.Context, string) error { return nil },
		},
	)

	err := svc.DeleteUserPhoto(context.Background(), uuid.Must(uuid.NewV7()), 10)
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument) {
		t.Fatalf("expected invalid argument, got: %v", err)
	}
}

func TestCreateUserConflictFromUniqueViolation(t *testing.T) {
	svc := NewService(
		&mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) {
				return domain.User{}, &pgconn.PgError{Code: "23505"}
			},
			getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
			updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
			deleteFn: func(context.Context, uuid.UUID) error { return nil },
			listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		},
		&mockUserPhotoRepo{
			setPrimaryFn: func(context.Context, uuid.UUID, string, string) error { return nil },
			replaceAllFn: func(context.Context, uuid.UUID, []domain.UserPhoto) error { return nil },
			replaceOneFn: func(context.Context, uuid.UUID, int64, domain.UserPhoto) error { return nil },
			upsertPosFn:  func(context.Context, uuid.UUID, domain.UserPhoto) error { return nil },
			getByIDFn:    func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:       func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
		},
		&mockS3Repo{
			presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
			presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
			deleteObjFn:  func(context.Context, string) error { return nil },
		},
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
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeConflict) {
		t.Fatalf("expected conflict error, got: %v", err)
	}
}
