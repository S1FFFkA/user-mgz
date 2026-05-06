package user

import (
	"context"
	"errors"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repo "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/jackc/pgx/v5"
)

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	userID, err := repo.NewUUIDv7()
	if err != nil {
		return domain.User{}, domain.InternalError(err)
	}
	user.ID = userID

	const userQuery = `
INSERT INTO users (
	id, first_name, last_name, email, birth_date, bio, toiler_score, alcohol_info, smoking_info, sex,
	height_cm, city_id, primary_photo_object_key, primary_photo_url
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING
	id, first_name, last_name, email, birth_date, bio, toiler_score, alcohol_info, smoking_info, sex,
	height_cm, city_id, primary_photo_object_key, primary_photo_url, created_at, updated_at`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	createdUser, err := repo.ScanUser(db.QueryRow(
		ctx,
		userQuery,
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
			return domain.User{}, domain.NotFoundError("resource not found", err)
		}
		return domain.User{}, domain.InternalError(err)
	}

	return createdUser, nil
}
