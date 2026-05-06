package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"go.uber.org/zap"
)

func (s *Server) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	userID, err := userproto.ParseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}

	current, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return nil, s.handleError("UpdateUser.GetUser", err, zap.String("user_id", req.GetUserId()))
	}

	updatedInput, err := userproto.MergeUpdateRequest(current, req)
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid update user request", err))
	}
	if err := ValidateUserForUpdate(updatedInput); err != nil {
		return nil, domain.ToGRPCStatus(err)
	}

	updated, err := s.users.UpdateUser(ctx, updatedInput)
	if err != nil {
		return nil, s.handleError("UpdateUser", err, zap.String("user_id", req.GetUserId()))
	}
	updated = s.enrichUserPhotoLinks(ctx, updated)
	return &userv1.UpdateUserResponse{User: userproto.ToProtoUser(updated)}, nil
}
