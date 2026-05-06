package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"go.uber.org/zap"
)

func (s *Server) ConfirmUserPhotoUpload(ctx context.Context, req *userv1.ConfirmUserPhotoUploadRequest) (*userv1.ConfirmUserPhotoUploadResponse, error) {
	userID, err := userproto.ParseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	photoType, err := userproto.FromProtoPhotoType(req.GetPhotoType())
	if err != nil {
		return nil, domain.ToGRPCStatus(err)
	}
	var extraPos *int16
	if req.GetExtraPosition() > 0 {
		v := int16(req.GetExtraPosition())
		extraPos = &v
	}
	user, err := s.photos.ConfirmPhotoUpload(ctx, domain.ConfirmPhotoUploadRequest{
		UserID:        userID,
		PhotoType:     photoType,
		ExtraPosition: extraPos,
		ObjectKey:     req.GetObjectKey(),
	})
	if err != nil {
		return nil, s.handleError("ConfirmUserPhotoUpload", err, zap.String("user_id", req.GetUserId()))
	}
	user = s.enrichUserPhotoLinks(ctx, user)
	return &userv1.ConfirmUserPhotoUploadResponse{User: userproto.ToProtoUser(user)}, nil
}
