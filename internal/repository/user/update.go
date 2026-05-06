package user

import (
	"context"
	"errors"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repo "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/jackc/pgx/v5"
)

func (r *UserRepository) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `
UPDATE users
SET
	first_name = $2,
	last_name = $3,
	email = $4,
	birth_date = $5,
	bio = $6,
	toiler_score = $7,
	alcohol_info = $8,
	smoking_info = $9,
	sex = $10,
	height_cm = $11,
	city_id = $12,
	primary_photo_object_key = $13,
	primary_photo_url = $14,
	updated_at = NOW()
WHERE id = $1
RETURNING
	id, first_name, last_name, email, birth_date, bio, toiler_score, alcohol_info, smoking_info, sex,
	height_cm, city_id, primary_photo_object_key, primary_photo_url, created_at, updated_at`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	updatedUser, err := repo.ScanUser(db.QueryRow(
		ctx,
		query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
		user.BirthDate,
		user.Bio,
		user.ToilerScore,
		user.AlcoholInfo,
		user.SmokingInfo,
		user.Sex,
		user.HeightCM,
		user.CityID,
		user.PrimaryPhotoObjectKey,
		user.PrimaryPhotoURL,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.NotFoundError("user not found", err)
		}
		return domain.User{}, domain.InternalError(err)
	}

	return updatedUser, nil
}
