package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) SetPrimaryPhoto(ctx context.Context, userID uuid.UUID, objectKey, url string) error {
	const query = `
UPDATE users
SET primary_photo_object_key = $2, primary_photo_url = $3
WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, userID, objectKey, url)
	if err != nil {
		return fmt.Errorf("set primary photo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("set primary photo: user not found")
	}
	return nil
}
