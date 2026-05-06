package userphoto

import (
	"context"
	"path/filepath"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) GetUserPhotoDownloadURL(ctx context.Context, req domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error) {
	if req.UserID == uuid.Nil {
		return domain.DownloadPhotoTicket{}, domain.InvalidArgumentError("invalid user_id", nil)
	}
	if req.PhotoType != domain.PhotoTypePrimary && req.PhotoType != domain.PhotoTypeExtra {
		return domain.DownloadPhotoTicket{}, domain.InvalidArgumentError("invalid photo_type", nil)
	}

	var objectKey string
	if req.PhotoType == domain.PhotoTypePrimary {
		user, err := s.userRepo.GetUserByID(ctx, req.UserID)
		if err != nil {
			s.log.Error("GetUserPhotoDownloadURL load user failed", zap.String("user_id", req.UserID.String()), zap.Error(err))
			return domain.DownloadPhotoTicket{}, err
		}
		objectKey = user.PrimaryPhotoObjectKey
	} else {
		if req.PhotoID == nil || *req.PhotoID <= 0 {
			return domain.DownloadPhotoTicket{}, domain.InvalidArgumentError("invalid photo_id", nil)
		}
		photo, err := s.userPhotoRepo.GetExtraPhotoByID(ctx, req.UserID, *req.PhotoID)
		if err != nil {
			s.log.Error("GetUserPhotoDownloadURL extra photo failed", zap.String("user_id", req.UserID.String()), zap.Int64("photo_id", *req.PhotoID), zap.Error(err))
			return domain.DownloadPhotoTicket{}, err
		}
		objectKey = photo.ObjectKey
	}

	fileName := filepath.Base(objectKey)
	downloadURL, err := s.storage.PresignGetURL(ctx, objectKey, fileName, false, domain.DefaultS3GetPresignTTL)
	if err != nil {
		s.log.Error("GetUserPhotoDownloadURL presign failed", zap.String("object_key", objectKey), zap.Error(err))
		return domain.DownloadPhotoTicket{}, domain.ServiceError("failed to generate download url", err)
	}

	return domain.DownloadPhotoTicket{
		DownloadURL:      downloadURL,
		ExpiresInSeconds: int64(domain.DefaultS3GetPresignTTL.Seconds()),
	}, nil
}
