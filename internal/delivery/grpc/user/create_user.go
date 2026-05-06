package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"go.uber.org/zap"
)

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user, err := userproto.FromCreateRequest(req)
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid create user request", err))
	}
	if err := ValidateUserForCreate(user); err != nil {
		return nil, domain.ToGRPCStatus(err)
	}
	created, err := s.users.CreateUser(ctx, user)
	if err != nil {
		return nil, s.handleError("CreateUser", err, zap.String("email", req.GetEmail()))
	}
	created = s.enrichUserPhotoLinks(ctx, created)
	return &userv1.CreateUserResponse{User: userproto.ToProtoUser(created)}, nil
}
