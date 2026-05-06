package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"go.uber.org/zap"
)

func (s *Server) DeleteUserPhoto(ctx context.Context, req *userv1.DeleteUserPhotoRequest) (*userv1.DeleteUserPhotoResponse, error) {
	userID, err := userproto.ParseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	if err = s.photos.DeleteUserPhoto(ctx, userID, req.GetPhotoId()); err != nil {
		return nil, s.handleError("DeleteUserPhoto", err,
			zap.String("user_id", req.GetUserId()),
			zap.Int64("photo_id", req.GetPhotoId()),
		)
	}
	return &userv1.DeleteUserPhotoResponse{}, nil
}
