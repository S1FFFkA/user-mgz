package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Server) enrichUserPhotoLinks(ctx context.Context, user domain.User) domain.User {
	if user.ID == uuid.Nil {
		return user
	}

	primary, err := s.photos.GetUserPhotoDownloadURL(ctx, domain.DownloadPhotoRequest{
		UserID:    user.ID,
		PhotoType: domain.PhotoTypePrimary,
	})
	if err == nil && primary.DownloadURL != "" {
		user.PrimaryPhotoURL = primary.DownloadURL
	} else if err != nil {
		s.logger.Warn("failed to enrich primary photo download url",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}

	for i := range user.ExtraPhotos {
		photoID := user.ExtraPhotos[i].ID
		if photoID <= 0 {
			continue
		}
		ticket, e := s.photos.GetUserPhotoDownloadURL(ctx, domain.DownloadPhotoRequest{
			UserID:    user.ID,
			PhotoType: domain.PhotoTypeExtra,
			PhotoID:   &photoID,
		})
		if e == nil && ticket.DownloadURL != "" {
			user.ExtraPhotos[i].URL = ticket.DownloadURL
			continue
		}
		if e != nil {
			s.logger.Warn("failed to enrich extra photo download url",
				zap.String("user_id", user.ID.String()),
				zap.Int64("photo_id", photoID),
				zap.Error(e),
			)
		}
	}

	return user
}
