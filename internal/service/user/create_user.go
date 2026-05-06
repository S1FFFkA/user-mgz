package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"go.uber.org/zap"
)

func (s *Service) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	s.log.Info("CreateUser started", zap.String("email", user.Email))
	var out domain.User
	err := s.tx.Do(ctx, func(ctx context.Context) error {
		created, err := s.userRepo.CreateUser(ctx, user)
		if err != nil {
			return err
		}
		out = created

		for _, p := range user.ExtraPhotos {
			if err = s.userPhotoRepo.InsertExtraPhoto(ctx, created.ID, p); err != nil {
				return err
			}
		}

		extras, err := s.userPhotoRepo.ListExtraPhotos(ctx, created.ID)
		if err != nil {
			return err
		}
		out.ExtraPhotos = extras
		return nil
	})
	if err != nil {
		s.log.Error("CreateUser failed", zap.Error(err))
		return domain.User{}, err
	}

	s.log.Info("CreateUser completed", zap.String("user_id", out.ID.String()))
	return out, nil
}
