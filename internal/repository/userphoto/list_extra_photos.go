package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repo "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
)

func (r *UserPhotoRepository) ListExtraPhotos(ctx context.Context, userID uuid.UUID) ([]domain.UserPhoto, error) {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	rows, err := db.Query(ctx, `
SELECT id, user_id, object_key, url, position, created_at
FROM user_photos
WHERE user_id = $1
ORDER BY position ASC`, userID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	photos, err := repo.CollectExtraPhotos(rows)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	return photos, nil
}
