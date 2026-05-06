package repository

import (
	"context"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error)
}

type UserPhotoRepositoryInterface interface {
	SetPrimaryPhoto(ctx context.Context, userID uuid.UUID, objectKey, url string) error
	DeleteAllExtraPhotos(ctx context.Context, userID uuid.UUID) error
	DeleteExtraPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error
	DeleteExtraPhotoByPosition(ctx context.Context, userID uuid.UUID, position int16) error
	InsertExtraPhoto(ctx context.Context, userID uuid.UUID, photo domain.UserPhoto) error
	GetExtraPhotoByID(ctx context.Context, userID uuid.UUID, photoID int64) (domain.UserPhoto, error)
	ListExtraPhotos(ctx context.Context, userID uuid.UUID) ([]domain.UserPhoto, error)
}

// S3RepositoryInterface — объектное хранилище (presign, удаление объектов).
type S3RepositoryInterface interface {
	PresignPutURL(ctx context.Context, objectKey, contentType string, contentLength int64, expiresIn time.Duration) (string, error)
	PresignGetURL(ctx context.Context, objectKey, fileName string, asAttachment bool, expiresIn time.Duration) (string, error)
	DeleteObject(ctx context.Context, objectKey string) error
}
