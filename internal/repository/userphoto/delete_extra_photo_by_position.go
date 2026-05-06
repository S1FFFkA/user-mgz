package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

func (r *UserPhotoRepository) DeleteExtraPhotoByPosition(ctx context.Context, userID uuid.UUID, position int16) error {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM user_photos WHERE user_id = $1 AND position = $2`, userID, position)
	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}
