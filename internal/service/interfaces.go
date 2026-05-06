package service

import (
	"context"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

// UserServiceInterface — сценарии профиля пользователя (БД users / списки).
type UserServiceInterface interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error)
}

// UserReader — загрузка пользователя с extra-фото (для сценариев userphoto после изменения фото).
type UserReader interface {
	GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error)
}

// UserPhotoServiceInterface — сценарии фото (БД user_photos + presigned URL).
type UserPhotoServiceInterface interface {
	GetUserPhotoUploadURL(ctx context.Context, req domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error)
	ConfirmPhotoUpload(ctx context.Context, req domain.ConfirmPhotoUploadRequest) (domain.User, error)
	DeleteUserPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error
	GetUserPhotoDownloadURL(ctx context.Context, req domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error)
}

// S3ObjectStorageInterface — объектное хранилище (реализация: internal/repository/s3).
type S3ObjectStorageInterface interface {
	PresignPutURL(ctx context.Context, objectKey, contentType string, contentLength int64, expiresIn time.Duration) (string, error)
	PresignGetURL(ctx context.Context, objectKey, fileName string, asAttachment bool, expiresIn time.Duration) (string, error)
	DeleteObject(ctx context.Context, objectKey string) error
}
