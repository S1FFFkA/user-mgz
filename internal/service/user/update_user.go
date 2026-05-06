package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	if user.ID == uuid.Nil {
		return domain.User{}, domain.InvalidArgumentError("invalid user_id", nil)
	}
	s.log.Info("UpdateUser started", zap.String("user_id", user.ID.String()))
	var out domain.User
	err := s.tx.Do(ctx, func(ctx context.Context) error {
		updated, err := s.userRepo.UpdateUser(ctx, user)
		if err != nil {
			return err
		}
		out = updated

		if len(user.ExtraPhotos) > 0 {
			if err = s.userPhotoRepo.DeleteAllExtraPhotos(ctx, user.ID); err != nil {
				return err
			}
			for _, p := range user.ExtraPhotos {
				if err = s.userPhotoRepo.InsertExtraPhoto(ctx, user.ID, p); err != nil {
					return err
				}
			}
		}

		extras, err := s.userPhotoRepo.ListExtraPhotos(ctx, user.ID)
		if err != nil {
			return err
		}
		out.ExtraPhotos = extras
		return nil
	})
	if err != nil {
		s.log.Error("UpdateUser failed", zap.String("user_id", user.ID.String()), zap.Error(err))
		return domain.User{}, err
	}

	s.log.Info("UpdateUser completed", zap.String("user_id", out.ID.String()))
	return out, nil
}
