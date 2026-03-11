package repository

import (
	"context"
	"fmt"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repocore "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
)

func (r *Repository) ListExtraPhotos(ctx context.Context, userID uuid.UUID) ([]domain.UserPhoto, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, user_id, object_key, url, position, created_at
FROM user_photos
WHERE user_id = $1
ORDER BY position ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list extra photos: %w", err)
	}
	defer rows.Close()

	return repocore.CollectExtraPhotos(rows)
}
