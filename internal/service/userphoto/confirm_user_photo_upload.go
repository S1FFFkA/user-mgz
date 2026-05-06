package userphoto

import (
	"context"
	"strings"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) ConfirmPhotoUpload(ctx context.Context, req domain.ConfirmPhotoUploadRequest) (domain.User, error) {
	if req.UserID == uuid.Nil || strings.TrimSpace(req.ObjectKey) == "" {
		return domain.User{}, domain.InvalidArgumentError("invalid photo upload confirm request", nil)
	}
	if req.PhotoType != domain.PhotoTypePrimary && req.PhotoType != domain.PhotoTypeExtra {
		return domain.User{}, domain.InvalidArgumentError("invalid photo_type", nil)
	}

	s.log.Info("ConfirmPhotoUpload", zap.String("user_id", req.UserID.String()), zap.String("photo_type", string(req.PhotoType)))
	if req.PhotoType == domain.PhotoTypePrimary {
		if err := s.userPhotoRepo.SetPrimaryPhoto(ctx, req.UserID, req.ObjectKey, req.ObjectKey); err != nil {
			s.log.Error("ConfirmPhotoUpload SetPrimaryPhoto failed", zap.Error(err))
			return domain.User{}, err
		}
	} else {
		if req.ExtraPosition == nil || *req.ExtraPosition < domain.MinExtraPhotoPosition || *req.ExtraPosition > domain.MaxExtraPhotoPosition {
			return domain.User{}, domain.InvalidArgumentError("invalid extra_position", nil)
		}
		if err := s.userPhotoRepo.DeleteExtraPhotoByPosition(ctx, req.UserID, *req.ExtraPosition); err != nil {
			s.log.Error("ConfirmPhotoUpload delete slot failed", zap.Error(err))
			return domain.User{}, err
		}
		photo := domain.UserPhoto{
			UserID:    req.UserID,
			ObjectKey: req.ObjectKey,
			URL:       req.ObjectKey,
			Position:  *req.ExtraPosition,
		}
		if err := s.userPhotoRepo.InsertExtraPhoto(ctx, req.UserID, photo); err != nil {
			s.log.Error("ConfirmPhotoUpload insert extra failed", zap.Error(err))
			return domain.User{}, err
		}
	}

	u, err := s.users.GetUser(ctx, req.UserID)
	if err != nil {
		s.log.Error("ConfirmPhotoUpload reload user failed", zap.Error(err))
	}
	return u, err
}
