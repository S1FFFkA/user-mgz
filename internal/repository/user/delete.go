package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

func (r *UserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const query = `DELETE FROM users WHERE id = $1`
	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	tag, err := db.Exec(ctx, query, userID)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFoundError("user not found", nil)
	}
	return nil
}
