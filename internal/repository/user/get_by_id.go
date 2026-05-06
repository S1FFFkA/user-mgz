package user

import (
	"context"
	"errors"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repo "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	const query = `
SELECT
	id, first_name, last_name, email, birth_date, bio, toiler_score, alcohol_info, smoking_info, sex,
	height_cm, city_id, primary_photo_object_key, primary_photo_url, created_at, updated_at
FROM users
WHERE id = $1`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	user, err := repo.ScanUser(db.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.NotFoundError("user not found", err)
		}
		return domain.User{}, domain.InternalError(err)
	}

	return user, nil
}
