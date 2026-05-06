package userphoto

import (
	"github.com/S1FFFkA/user-mgz/internal/repository"
	svc "github.com/S1FFFkA/user-mgz/internal/service"
	"go.uber.org/zap"
)

// Service — загрузка/скачивание фото и фиксация в БД.
type Service struct {
	userRepo      repository.UserRepositoryInterface
	userPhotoRepo repository.UserPhotoRepositoryInterface
	storage       svc.S3ObjectStorageInterface
	users         svc.UserReader
	log           *zap.Logger
}

func NewService(
	userRepo repository.UserRepositoryInterface,
	userPhotoRepo repository.UserPhotoRepositoryInterface,
	storage svc.S3ObjectStorageInterface,
	users svc.UserReader,
	log *zap.Logger,
) *Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &Service{
		userRepo:      userRepo,
		userPhotoRepo: userPhotoRepo,
		storage:       storage,
		users:         users,
		log:           log,
	}
}

var _ svc.UserPhotoServiceInterface = (*Service)(nil)
