package userphoto

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

func (r *UserPhotoRepository) InsertExtraPhoto(ctx context.Context, userID uuid.UUID, photo domain.UserPhoto) error {
	const query = `
INSERT INTO user_photos (user_id, object_key, url, position)
VALUES ($1, $2, $3, $4)`
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	if _, err := db.Exec(ctx, query, userID, photo.ObjectKey, photo.URL, photo.Position); err != nil {
		return domain.InternalError(err)
	}
	return nil
}
