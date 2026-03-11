package repository

import (
	"context"
	"fmt"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

func (r *Repository) UpsertExtraPhotoByPosition(ctx context.Context, userID uuid.UUID, photo domain.UserPhoto) error {
	const query = `
INSERT INTO user_photos (user_id, object_key, url, position)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, position)
DO UPDATE SET object_key = EXCLUDED.object_key, url = EXCLUDED.url`

	_, err := r.pool.Exec(ctx, query, userID, photo.ObjectKey, photo.URL, photo.Position)
	if err != nil {
		return fmt.Errorf("upsert extra photo by position: %w", err)
	}
	return nil
}
