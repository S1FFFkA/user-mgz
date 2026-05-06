package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repo "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/jackc/pgx/v5"
)

const defaultUsersLimit = 20

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error) {
	if limit <= 0 {
		limit = defaultUsersLimit
	}
	if offset < 0 {
		offset = 0
	}

	const queryNoCity = `
SELECT
	id, first_name, last_name, email, birth_date, bio, toiler_score, alcohol_info, smoking_info, sex,
	height_cm, city_id, primary_photo_object_key, primary_photo_url, created_at, updated_at
FROM users
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2`

	const queryByCity = `
SELECT
	id, first_name, last_name, email, birth_date, bio, toiler_score, alcohol_info, smoking_info, sex,
	height_cm, city_id, primary_photo_object_key, primary_photo_url, created_at, updated_at
FROM users
WHERE city_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	var err error
	var rows pgx.Rows

	if cityID == nil {
		rows, err = db.Query(ctx, queryNoCity, limit, offset)
	} else {
		rows, err = db.Query(ctx, queryByCity, *cityID, limit, offset)
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	result := make([]domain.User, 0, limit)
	for rows.Next() {
		user, scanErr := repo.ScanUser(rows)
		if scanErr != nil {
			return nil, domain.InternalError(scanErr)
		}
		result = append(result, user)
	}
	if err = rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}

	return result, nil
}
