package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"go.uber.org/zap"
)

func (s *Server) GetUserPhotoUploadUrl(ctx context.Context, req *userv1.GetUserPhotoUploadUrlRequest) (*userv1.GetUserPhotoUploadUrlResponse, error) {
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

	uploadReq := domain.UploadPhotoRequest{
		UserID:        userID,
		PhotoType:     photoType,
		ExtraPosition: extraPos,
		ContentType:   req.GetContentType(),
		ContentLength: req.GetContentLength(),
	}
	if err := ValidateUploadPhotoRequest(uploadReq); err != nil {
		return nil, domain.ToGRPCStatus(err)
	}
	ticket, err := s.photos.GetUserPhotoUploadURL(ctx, uploadReq)
	if err != nil {
		return nil, s.handleError("GetUserPhotoUploadUrl", err, zap.String("user_id", req.GetUserId()))
	}

	return &userv1.GetUserPhotoUploadUrlResponse{
		ObjectKey:        ticket.ObjectKey,
		UploadUrl:        ticket.UploadURL,
		ExpiresInSeconds: ticket.ExpiresInSeconds,
	}, nil
}
