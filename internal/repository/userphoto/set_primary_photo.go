package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

func (r *UserPhotoRepository) SetPrimaryPhoto(ctx context.Context, userID uuid.UUID, objectKey, url string) error {
	const query = `
UPDATE users
SET primary_photo_object_key = $2, primary_photo_url = $3, updated_at = NOW()
WHERE id = $1`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	tag, err := db.Exec(ctx, query, userID, objectKey, url)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFoundError("user not found", nil)
	}
	return nil
}
