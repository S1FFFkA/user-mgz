package user

import (
	"github.com/S1FFFkA/user-mgz/internal/service"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"go.uber.org/zap"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	users  service.UserServiceInterface
	photos service.UserPhotoServiceInterface
	logger *zap.Logger
}

func NewServer(users service.UserServiceInterface, photos service.UserPhotoServiceInterface, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{
		users:  users,
		photos: photos,
		logger: logger,
	}
}
