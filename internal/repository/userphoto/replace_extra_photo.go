package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) ReplaceExtraPhoto(ctx context.Context, userID uuid.UUID, photoID int64, photo domain.UserPhoto) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx replace extra photo: %w", err)
	}
	defer tx.Rollback(ctx)

	var oldPosition int16
	const selectQuery = `
SELECT position
FROM user_photos
WHERE id = $1 AND user_id = $2
FOR UPDATE`
	if err = tx.QueryRow(ctx, selectQuery, photoID, userID).Scan(&oldPosition); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("replace extra photo: photo not found")
		}
		return fmt.Errorf("replace extra photo select old: %w", err)
	}

	const deleteQuery = `DELETE FROM user_photos WHERE id = $1 AND user_id = $2`
	if _, err = tx.Exec(ctx, deleteQuery, photoID, userID); err != nil {
		return fmt.Errorf("replace extra photo delete old: %w", err)
	}

	position := photo.Position
	if position == 0 {
		position = oldPosition
	}

	const insertQuery = `
INSERT INTO user_photos (user_id, object_key, url, position)
VALUES ($1, $2, $3, $4)`
	if _, err = tx.Exec(ctx, insertQuery, userID, photo.ObjectKey, photo.URL, position); err != nil {
		return fmt.Errorf("replace extra photo insert new: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace extra photo: %w", err)
	}

	return nil
}
