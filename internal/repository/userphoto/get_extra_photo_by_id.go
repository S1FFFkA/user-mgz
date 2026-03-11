package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repocore "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetExtraPhotoByID(ctx context.Context, userID uuid.UUID, photoID int64) (domain.UserPhoto, error) {
	const query = `
SELECT id, user_id, object_key, url, position, created_at
FROM user_photos
WHERE id = $1 AND user_id = $2`

	photo, err := repocore.ScanUserPhoto(r.pool.QueryRow(ctx, query, photoID, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserPhoto{}, fmt.Errorf("extra photo not found: %w", err)
		}
		return domain.UserPhoto{}, fmt.Errorf("get extra photo by id: %w", err)
	}
	return photo, nil
}
