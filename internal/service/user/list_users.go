package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"go.uber.org/zap"
)

func (s *Service) ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error) {
	if limit <= 0 {
		limit = domain.DefaultUsersListLimit
	}
	if limit > domain.MaxUsersListLimit {
		limit = domain.MaxUsersListLimit
	}
	if offset < 0 {
		offset = 0
	}
	if cityID != nil && *cityID <= 0 {
		return nil, domain.InvalidArgumentError("invalid city_id", nil)
	}

	users, err := s.userRepo.ListUsers(ctx, limit, offset, cityID)
	if err != nil {
		s.log.Error("ListUsers failed", zap.Int32("limit", limit), zap.Int32("offset", offset), zap.Error(err))
		return nil, err
	}

	for i := range users {
		users[i].ExtraPhotos, err = s.userPhotoRepo.ListExtraPhotos(ctx, users[i].ID)
		if err != nil {
			s.log.Error("ListUsers extra photos failed", zap.String("user_id", users[i].ID.String()), zap.Error(err))
			return nil, err
		}
	}

	return users, nil
}
