package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"go.uber.org/zap"
)

func (s *Service) GetUserPhotoUploadURL(ctx context.Context, req domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error) {
	s.log.Info("GetUserPhotoUploadURL", zap.String("user_id", req.UserID.String()), zap.String("photo_type", string(req.PhotoType)))
	ext := domain.PhotoFileExtensionFromContentType(req.ContentType)
	objectKey := domain.UserPhotoObjectKey(req.UserID, req.PhotoType, req.ExtraPosition, ext)

	uploadURL, err := s.storage.PresignPutURL(ctx, objectKey, req.ContentType, req.ContentLength, domain.DefaultS3PutPresignTTL)
	if err != nil {
		s.log.Error("GetUserPhotoUploadURL presign failed", zap.String("object_key", objectKey), zap.Error(err))
		return domain.UploadPhotoTicket{}, domain.ServiceError("failed to generate upload url", err)
	}

	return domain.UploadPhotoTicket{
		ObjectKey:        objectKey,
		UploadURL:        uploadURL,
		ExpiresInSeconds: int64(domain.DefaultS3PutPresignTTL.Seconds()),
	}, nil
}
