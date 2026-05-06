package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"go.uber.org/zap"
)

func (s *Server) GetUserPhotoDownloadUrl(ctx context.Context, req *userv1.GetUserPhotoDownloadUrlRequest) (*userv1.GetUserPhotoDownloadUrlResponse, error) {
	userID, err := userproto.ParseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	photoType, err := userproto.FromProtoPhotoType(req.GetPhotoType())
	if err != nil {
		return nil, domain.ToGRPCStatus(err)
	}
	var photoID *int64
	if req.GetPhotoId() > 0 {
		v := req.GetPhotoId()
		photoID = &v
	}

	ticket, err := s.photos.GetUserPhotoDownloadURL(ctx, domain.DownloadPhotoRequest{
		UserID:    userID,
		PhotoType: photoType,
		PhotoID:   photoID,
	})
	if err != nil {
		return nil, s.handleError("GetUserPhotoDownloadUrl", err, zap.String("user_id", req.GetUserId()))
	}

	return &userv1.GetUserPhotoDownloadUrlResponse{
		DownloadUrl:      ticket.DownloadURL,
		ExpiresInSeconds: ticket.ExpiresInSeconds,
	}, nil
}
