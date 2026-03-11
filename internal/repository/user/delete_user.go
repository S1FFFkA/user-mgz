package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const query = `DELETE FROM users WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete user: no rows affected")
	}
	return nil
}
