package repository

import (
	"context"
	"fmt"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repocore "github.com/S1FFFkA/user-mgz/internal/repository"
)

func (r *Repository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	userID, err := repocore.NewUUIDv7()
	if err != nil {
		return domain.User{}, fmt.Errorf("generate uuidv7 for user: %w", err)
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

	createdUser, err := repocore.ScanUser(r.pool.QueryRow(
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
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}
