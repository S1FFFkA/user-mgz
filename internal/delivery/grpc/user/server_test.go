package user

import (
	"context"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/S1FFFkA/user-mgz/internal/service"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockUseCase struct {
	createFn        func(ctx context.Context, user domain.User) (domain.User, error)
	getFn           func(ctx context.Context, userID uuid.UUID) (domain.User, error)
	updateFn        func(ctx context.Context, user domain.User) (domain.User, error)
	deleteFn        func(ctx context.Context, userID uuid.UUID) error
	listFn          func(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error)
	uploadURLFn     func(ctx context.Context, req domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error)
	confirmUploadFn func(ctx context.Context, req domain.ConfirmPhotoUploadRequest) (domain.User, error)
	deletePhotoFn   func(ctx context.Context, userID uuid.UUID, photoID int64) error
	downloadURLFn   func(ctx context.Context, req domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error)
}

var _ service.UserServiceInterface = (*mockUseCase)(nil)
var _ service.UserPhotoServiceInterface = (*mockUseCase)(nil)

func (m *mockUseCase) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.createFn(ctx, user)
}
func (m *mockUseCase) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	return m.getFn(ctx, userID)
}
func (m *mockUseCase) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.updateFn(ctx, user)
}
func (m *mockUseCase) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return m.deleteFn(ctx, userID)
}
func (m *mockUseCase) ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error) {
	return m.listFn(ctx, limit, offset, cityID)
}
func (m *mockUseCase) GetUserPhotoUploadURL(ctx context.Context, req domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error) {
	return m.uploadURLFn(ctx, req)
}
func (m *mockUseCase) ConfirmPhotoUpload(ctx context.Context, req domain.ConfirmPhotoUploadRequest) (domain.User, error) {
	return m.confirmUploadFn(ctx, req)
}
func (m *mockUseCase) DeleteUserPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error {
	return m.deletePhotoFn(ctx, userID, photoID)
}
func (m *mockUseCase) GetUserPhotoDownloadURL(ctx context.Context, req domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error) {
	return m.downloadURLFn(ctx, req)
}

func TestCreateUserInvalidBirthDate(t *testing.T) {
	m := &mockUseCase{
		createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
		updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		uploadURLFn: func(context.Context, domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error) {
			return domain.UploadPhotoTicket{}, nil
		},
		confirmUploadFn: func(context.Context, domain.ConfirmPhotoUploadRequest) (domain.User, error) {
			return domain.User{}, nil
		},
		deletePhotoFn: func(context.Context, uuid.UUID, int64) error { return nil },
		downloadURLFn: func(context.Context, domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error) {
			return domain.DownloadPhotoTicket{}, nil
		},
	}
	s := NewServer(m, m, nil)

	_, err := s.CreateUser(context.Background(), &userv1.CreateUserRequest{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             "bad-date",
		ToilerScore:           7,
		Sex:                   userv1.Sex_SEX_MALE,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoUrl:       "u",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v (err=%v)", status.Code(err), err)
	}
}

func TestGetUserNotFoundMapsToNotFound(t *testing.T) {
	m := &mockUseCase{
		createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		getFn: func(context.Context, uuid.UUID) (domain.User, error) {
			return domain.User{}, domain.NotFoundError("user not found", nil)
		},
		updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		uploadURLFn: func(context.Context, domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error) {
			return domain.UploadPhotoTicket{}, nil
		},
		confirmUploadFn: func(context.Context, domain.ConfirmPhotoUploadRequest) (domain.User, error) {
			return domain.User{}, nil
		},
		deletePhotoFn: func(context.Context, uuid.UUID, int64) error { return nil },
		downloadURLFn: func(context.Context, domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error) {
			return domain.DownloadPhotoTicket{}, nil
		},
	}
	s := NewServer(m, m, nil)

	uid := uuid.Must(uuid.NewV7())
	_, err := s.GetUser(context.Background(), &userv1.GetUserRequest{UserId: uid.String()})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v (err=%v)", status.Code(err), err)
	}
}

func TestGetUserPhotoUploadUrlInvalidPhotoType(t *testing.T) {
	m := &mockUseCase{
		createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		getFn:    func(context.Context, uuid.UUID) (domain.User, error) { return domain.User{}, nil },
		updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		uploadURLFn: func(context.Context, domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error) {
			return domain.UploadPhotoTicket{}, nil
		},
		confirmUploadFn: func(context.Context, domain.ConfirmPhotoUploadRequest) (domain.User, error) {
			return domain.User{}, nil
		},
		deletePhotoFn: func(context.Context, uuid.UUID, int64) error { return nil },
		downloadURLFn: func(context.Context, domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error) {
			return domain.DownloadPhotoTicket{}, nil
		},
	}
	s := NewServer(m, m, nil)

	uid := uuid.Must(uuid.NewV7())
	_, err := s.GetUserPhotoUploadUrl(context.Background(), &userv1.GetUserPhotoUploadUrlRequest{
		UserId:        uid.String(),
		PhotoType:     userv1.PhotoType_PHOTO_TYPE_UNSPECIFIED,
		ContentType:   "image/jpeg",
		ContentLength: 100,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v (err=%v)", status.Code(err), err)
	}
}

func TestGetUserSuccess(t *testing.T) {
	uid := uuid.Must(uuid.NewV7())
	now := time.Now()
	m := &mockUseCase{
		createFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		getFn: func(context.Context, uuid.UUID) (domain.User, error) {
			return domain.User{
				ID:                    uid,
				FirstName:             "A",
				LastName:              "B",
				Email:                 "a@b.com",
				BirthDate:             time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC),
				ToilerScore:           8,
				Sex:                   domain.SexMale,
				PrimaryPhotoObjectKey: "k",
				PrimaryPhotoURL:       "u",
				CreatedAt:             now,
				UpdatedAt:             now,
			}, nil
		},
		updateFn: func(context.Context, domain.User) (domain.User, error) { return domain.User{}, nil },
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		listFn:   func(context.Context, int32, int32, *int64) ([]domain.User, error) { return nil, nil },
		uploadURLFn: func(context.Context, domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error) {
			return domain.UploadPhotoTicket{}, nil
		},
		confirmUploadFn: func(context.Context, domain.ConfirmPhotoUploadRequest) (domain.User, error) {
			return domain.User{}, nil
		},
		deletePhotoFn: func(context.Context, uuid.UUID, int64) error { return nil },
		downloadURLFn: func(context.Context, domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error) {
			return domain.DownloadPhotoTicket{}, nil
		},
	}
	s := NewServer(m, m, nil)

	resp, err := s.GetUser(context.Background(), &userv1.GetUserRequest{UserId: uid.String()})
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if resp.GetUser().GetId() != uid.String() {
		t.Fatalf("unexpected user id")
	}
}
