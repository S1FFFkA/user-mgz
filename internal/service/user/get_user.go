package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	if userID == uuid.Nil {
		return domain.User{}, domain.InvalidArgumentError("invalid user_id", nil)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.log.Error("GetUser repository failed", zap.String("user_id", userID.String()), zap.Error(err))
		return domain.User{}, err
	}

	user.ExtraPhotos, err = s.userPhotoRepo.ListExtraPhotos(ctx, user.ID)
	if err != nil {
		s.log.Error("GetUser extra photos failed", zap.String("user_id", userID.String()), zap.Error(err))
		return domain.User{}, err
	}

	return user, nil
}
