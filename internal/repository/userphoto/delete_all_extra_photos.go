package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

func (r *UserPhotoRepository) DeleteAllExtraPhotos(ctx context.Context, userID uuid.UUID) error {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM user_photos WHERE user_id = $1`, userID)
	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}
