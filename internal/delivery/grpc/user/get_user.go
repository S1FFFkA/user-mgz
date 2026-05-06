package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"go.uber.org/zap"
)

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	userID, err := userproto.ParseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return nil, s.handleError("GetUser", err, zap.String("user_id", req.GetUserId()))
	}
	user = s.enrichUserPhotoLinks(ctx, user)
	return &userv1.GetUserResponse{User: userproto.ToProtoUser(user)}, nil
}
