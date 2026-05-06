package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"go.uber.org/zap"
)

func (s *Server) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	userID, err := userproto.ParseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	if err = s.users.DeleteUser(ctx, userID); err != nil {
		return nil, s.handleError("DeleteUser", err, zap.String("user_id", req.GetUserId()))
	}
	return &userv1.DeleteUserResponse{}, nil
}
