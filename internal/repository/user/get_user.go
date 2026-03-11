package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repocore "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	const query = `
SELECT
	id, first_name, last_name, email, birth_date, bio, toiler_score, alcohol_info, smoking_info, sex,
	height_cm, city_id, primary_photo_object_key, primary_photo_url, created_at, updated_at
FROM users
WHERE id = $1`

	user, err := repocore.ScanUser(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user not found: %w", err)
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}
