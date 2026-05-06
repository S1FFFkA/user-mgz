package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

func (r *UserPhotoRepository) DeleteExtraPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error {
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	tag, err := db.Exec(ctx, `DELETE FROM user_photos WHERE id = $1 AND user_id = $2`, photoID, userID)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFoundError("extra photo not found", nil)
	}
	return nil
}
