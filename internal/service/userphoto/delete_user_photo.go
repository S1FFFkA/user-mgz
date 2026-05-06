package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) DeleteUserPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error {
	if userID == uuid.Nil {
		return domain.InvalidArgumentError("invalid user_id", nil)
	}
	if photoID <= 0 {
		return domain.InvalidArgumentError("invalid photo_id", nil)
	}

	// В API удаление фото сейчас относится к extra-фото (user_photos). Primary фото удаляется/заменяется отдельным сценарием.
	photo, err := s.userPhotoRepo.GetExtraPhotoByID(ctx, userID, photoID)
	if err != nil {
		return err
	}

	// Сначала убираем из профиля (БД), чтобы пользователь сразу перестал видеть фото.
	if err := s.userPhotoRepo.DeleteExtraPhoto(ctx, userID, photoID); err != nil {
		return err
	}

	// Объект в S3 удаляем best-effort: если не получилось — фото в профиле уже удалено, это “утечка” в storage,
	// но не ломает UX. Ошибку логируем.
	if photo.ObjectKey != "" {
		if err := s.storage.DeleteObject(ctx, photo.ObjectKey); err != nil {
			s.log.Warn("DeleteUserPhoto storage delete failed",
				zap.String("user_id", userID.String()),
				zap.Int64("photo_id", photoID),
				zap.String("object_key", photo.ObjectKey),
				zap.Error(err),
			)
		}
	}
	return nil
}
