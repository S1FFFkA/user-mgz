package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return domain.InvalidArgumentError("invalid user_id", nil)
	}
	s.log.Info("DeleteUser", zap.String("user_id", userID.String()))
	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		s.log.Error("DeleteUser failed", zap.String("user_id", userID.String()), zap.Error(err))
		return err
	}
	return nil
}
