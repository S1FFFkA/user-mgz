package user

import (
	"github.com/S1FFFkA/user-mgz/internal/repository"
	svc "github.com/S1FFFkA/user-mgz/internal/service"
	"go.uber.org/zap"
)

// Service — профиль пользователя и списки (БД users + связка с user_photos).
type Service struct {
	userRepo      repository.UserRepositoryInterface
	userPhotoRepo repository.UserPhotoRepositoryInterface
	tx            svc.TxManagerInterface
	log           *zap.Logger
}

func NewService(
	userRepo repository.UserRepositoryInterface,
	userPhotoRepo repository.UserPhotoRepositoryInterface,
	tx svc.TxManagerInterface,
	log *zap.Logger,
) *Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &Service{
		userRepo:      userRepo,
		userPhotoRepo: userPhotoRepo,
		tx:            tx,
		log:           log,
	}
}

var _ svc.UserServiceInterface = (*Service)(nil)
var _ svc.UserReader = (*Service)(nil)
