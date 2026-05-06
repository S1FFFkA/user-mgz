package userphoto

import (
	"context"
	"errors"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repo "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *UserPhotoRepository) GetExtraPhotoByID(ctx context.Context, userID uuid.UUID, photoID int64) (domain.UserPhoto, error) {
	const query = `
SELECT id, user_id, object_key, url, position, created_at
FROM user_photos
WHERE id = $1 AND user_id = $2`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	photo, err := repo.ScanUserPhoto(db.QueryRow(ctx, query, photoID, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserPhoto{}, domain.NotFoundError("extra photo not found", err)
		}
		return domain.UserPhoto{}, domain.InternalError(err)
	}
	return photo, nil
}
