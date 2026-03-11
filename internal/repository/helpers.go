package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RowScanner interface {
	Scan(dest ...any) error
}

type RowIterator interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

func NewUUIDv7() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func ScanUser(scanner RowScanner) (domain.User, error) {
	var (
		user             domain.User
		bio              *string
		alcoholInfo      *string
		smokingInfo      *string
		heightCM         *int16
		cityID           *int64
		createdAt        time.Time
		updatedAt        time.Time
		primaryObjectKey string
		primaryPhotoURL  string
	)

	err := scanner.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.BirthDate,
		&bio,
		&user.ToilerScore,
		&alcoholInfo,
		&smokingInfo,
		&user.Sex,
		&heightCM,
		&cityID,
		&primaryObjectKey,
		&primaryPhotoURL,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}

	user.Bio = bio
	user.AlcoholInfo = alcoholInfo
	user.SmokingInfo = smokingInfo
	user.HeightCM = heightCM
	user.CityID = cityID
	user.PrimaryPhotoObjectKey = primaryObjectKey
	user.PrimaryPhotoURL = primaryPhotoURL
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return user, nil
}

func ScanUserPhoto(scanner RowScanner) (domain.UserPhoto, error) {
	var photo domain.UserPhoto
	err := scanner.Scan(
		&photo.ID,
		&photo.UserID,
		&photo.ObjectKey,
		&photo.URL,
		&photo.Position,
		&photo.CreatedAt,
	)
	if err != nil {
		return domain.UserPhoto{}, fmt.Errorf("scan user photo: %w", err)
	}
	return photo, nil
}

func ReplaceExtraPhotosTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, photos []domain.UserPhoto) error {
	if _, err := tx.Exec(ctx, `DELETE FROM user_photos WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete old extra photos: %w", err)
	}
	return InsertExtraPhotosTx(ctx, tx, userID, photos)
}

func InsertExtraPhotosTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, photos []domain.UserPhoto) error {
	if len(photos) == 0 {
		return nil
	}

	const query = `
INSERT INTO user_photos (user_id, object_key, url, position)
VALUES ($1, $2, $3, $4)`

	for _, photo := range photos {
		if _, err := tx.Exec(ctx, query, userID, photo.ObjectKey, photo.URL, photo.Position); err != nil {
			return fmt.Errorf("insert extra photo: %w", err)
		}
	}
	return nil
}

func CollectExtraPhotos(rows RowIterator) ([]domain.UserPhoto, error) {
	result := make([]domain.UserPhoto, 0)
	for rows.Next() {
		photo, err := ScanUserPhoto(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, photo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("collect extra photos rows: %w", err)
	}
	return result, nil
}
