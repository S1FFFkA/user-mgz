package userphoto

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	usersvc "github.com/S1FFFkA/user-mgz/internal/service/user"
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

type mockS3 struct {
	presignPutFn func(ctx context.Context, objectKey, contentType string, contentLength int64, expiresIn time.Duration) (string, error)
	presignGetFn func(ctx context.Context, objectKey, fileName string, asAttachment bool, expiresIn time.Duration) (string, error)
	deleteObjFn  func(ctx context.Context, objectKey string) error
}

func (m *mockS3) PresignPutURL(ctx context.Context, objectKey, contentType string, contentLength int64, expiresIn time.Duration) (string, error) {
	return m.presignPutFn(ctx, objectKey, contentType, contentLength, expiresIn)
}
func (m *mockS3) PresignGetURL(ctx context.Context, objectKey, fileName string, asAttachment bool, expiresIn time.Duration) (string, error) {
	return m.presignGetFn(ctx, objectKey, fileName, asAttachment, expiresIn)
}
func (m *mockS3) DeleteObject(ctx context.Context, objectKey string) error {
	return m.deleteObjFn(ctx, objectKey)
}

func stubUserRepos() (*mockUserRepo, *mockUserPhotoRepo) {
	return &mockUserRepo{
			createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
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
			insertFn:           func(context.Context, uuid.UUID, domain.UserPhoto) error { return nil },
			getByIDFn:          func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
			listFn:             func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
		}
}

func TestGetUserPhotoUploadURLSuccess(t *testing.T) {
	var gotObjectKey string
	userID := uuid.Must(uuid.NewV7())
	ur, pr := stubUserRepos()
	userReader := usersvc.NewService(ur, pr, testutil.PassthroughTxManager{}, zap.NewNop())

	svc := NewService(ur, pr, &mockS3{
		presignPutFn: func(_ context.Context, objectKey, _ string, _ int64, _ time.Duration) (string, error) {
			gotObjectKey = objectKey
			return "https://upload", nil
		},
		presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
		deleteObjFn:  func(context.Context, string) error { return nil },
	}, userReader, zap.NewNop())

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

func TestConfirmUserPhotoUploadExtraDeletesSlotThenInserts(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	var deletePos, insertCalls int

	ur := &mockUserRepo{
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
	}
	pr := &mockUserPhotoRepo{
		setPrimaryFn:  func(context.Context, uuid.UUID, string, string) error { return nil },
		deleteAllFn:   func(context.Context, uuid.UUID) error { return nil },
		deletePhotoFn: func(context.Context, uuid.UUID, int64) error { return nil },
		deleteByPositionFn: func(_ context.Context, gotUserID uuid.UUID, pos int16) error {
			deletePos++
			if gotUserID != userID || pos != 1 {
				t.Fatalf("unexpected delete by position args")
			}
			return nil
		},
		insertFn: func(_ context.Context, gotUserID uuid.UUID, photo domain.UserPhoto) error {
			insertCalls++
			if gotUserID != userID || photo.Position != 1 {
				t.Fatalf("unexpected insert args")
			}
			return nil
		},
		getByIDFn: func(context.Context, uuid.UUID, int64) (domain.UserPhoto, error) { return domain.UserPhoto{}, nil },
		listFn:    func(context.Context, uuid.UUID) ([]domain.UserPhoto, error) { return nil, nil },
	}
	userReader := usersvc.NewService(ur, pr, testutil.PassthroughTxManager{}, zap.NewNop())

	svc := NewService(ur, pr, &mockS3{
		presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
		presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
		deleteObjFn:  func(context.Context, string) error { return nil },
	}, userReader, zap.NewNop())

	pos := int16(1)
	_, err := svc.ConfirmPhotoUpload(context.Background(), domain.ConfirmPhotoUploadRequest{
		UserID:        userID,
		PhotoType:     domain.PhotoTypeExtra,
		ExtraPosition: &pos,
		ObjectKey:     "users/u/extra/1/new.jpg",
	})
	if err != nil {
		t.Fatalf("ConfirmPhotoUpload error: %v", err)
	}
	if deletePos != 1 || insertCalls != 1 {
		t.Fatalf("expected 1 delete-by-position and 1 insert, got delete=%d insert=%d", deletePos, insertCalls)
	}
}

func TestDeleteUserPhoto_DeletesExtraPhotoAndObject(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	var (
		gotUserID    uuid.UUID
		gotPhotoID   int64
		gotObjectKey string
	)

	ur, pr := stubUserRepos()
	pr.getByIDFn = func(ctx context.Context, uid uuid.UUID, pid int64) (domain.UserPhoto, error) {
		gotUserID, gotPhotoID = uid, pid
		return domain.UserPhoto{
			ID:        pid,
			UserID:    uid,
			ObjectKey: "users/u/extra/1/p.png",
			URL:       "users/u/extra/1/p.png",
			Position:  1,
		}, nil
	}
	pr.deletePhotoFn = func(ctx context.Context, uid uuid.UUID, pid int64) error {
		gotUserID, gotPhotoID = uid, pid
		return nil
	}

	userReader := usersvc.NewService(ur, pr, testutil.PassthroughTxManager{}, zap.NewNop())
	svc := NewService(ur, pr, &mockS3{
		presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
		presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
		deleteObjFn: func(ctx context.Context, key string) error {
			gotObjectKey = key
			return nil
		},
	}, userReader, zap.NewNop())

	if err := svc.DeleteUserPhoto(context.Background(), userID, 10); err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}
	if gotUserID != userID || gotPhotoID != 10 {
		t.Fatalf("expected repo called with userID/photoID, got userID=%s photoID=%d", gotUserID, gotPhotoID)
	}
	if gotObjectKey != "users/u/extra/1/p.png" {
		t.Fatalf("expected DeleteObject called, got key=%q", gotObjectKey)
	}
}

func TestGetUserPhotoDownloadURLInvalidUserID(t *testing.T) {
	ur, pr := stubUserRepos()
	userReader := usersvc.NewService(ur, pr, testutil.PassthroughTxManager{}, zap.NewNop())
	svc := NewService(ur, pr, &mockS3{
		presignPutFn: func(context.Context, string, string, int64, time.Duration) (string, error) { return "", nil },
		presignGetFn: func(context.Context, string, string, bool, time.Duration) (string, error) { return "", nil },
		deleteObjFn:  func(context.Context, string) error { return nil },
	}, userReader, zap.NewNop())

	_, err := svc.GetUserPhotoDownloadURL(context.Background(), domain.DownloadPhotoRequest{
		UserID:    uuid.Nil,
		PhotoType: domain.PhotoTypePrimary,
	})
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument) {
		t.Fatalf("expected invalid argument, got: %v", err)
	}
}
