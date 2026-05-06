package user

import (
	"strings"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

// Проверки входа на границе gRPC: контракт API и инварианты, общие для транспорта.
// Сервисный слой предполагает уже согласованный доменный запрос.

func ValidateUserForCreate(user domain.User) error {
	if user.FirstName == "" || len(user.FirstName) > domain.MaxFirstNameLength {
		return domain.InvalidArgumentError("invalid first_name", nil)
	}
	if user.LastName == "" || len(user.LastName) > domain.MaxLastNameLength {
		return domain.InvalidArgumentError("invalid last_name", nil)
	}
	if user.Email == "" || len(user.Email) > domain.MaxEmailLength || !strings.Contains(user.Email, "@") {
		return domain.InvalidArgumentError("invalid email", nil)
	}
	if user.BirthDate.IsZero() {
		return domain.InvalidArgumentError("invalid birth_date", nil)
	}
	if user.ToilerScore < domain.MinToilerScore || user.ToilerScore > domain.MaxToilerScore {
		return domain.InvalidArgumentError("invalid toiler_score", nil)
	}
	if user.Sex != domain.SexMale && user.Sex != domain.SexFemale {
		return domain.InvalidArgumentError("invalid sex", nil)
	}
	if user.HeightCM != nil && (*user.HeightCM < 100 || *user.HeightCM > 260) {
		return domain.InvalidArgumentError("invalid height_cm", nil)
	}
	if strings.TrimSpace(user.PrimaryPhotoObjectKey) == "" || strings.TrimSpace(user.PrimaryPhotoURL) == "" {
		return domain.InvalidArgumentError("primary photo is required", nil)
	}
	for _, photo := range user.ExtraPhotos {
		if photo.Position < domain.MinExtraPhotoPosition || photo.Position > domain.MaxExtraPhotoPosition {
			return domain.InvalidArgumentError("invalid extra photo position", nil)
		}
	}
	return nil
}

func ValidateUserForUpdate(user domain.User) error {
	if user.FirstName != "" && len(user.FirstName) > domain.MaxFirstNameLength {
		return domain.InvalidArgumentError("invalid first_name", nil)
	}
	if user.LastName != "" && len(user.LastName) > domain.MaxLastNameLength {
		return domain.InvalidArgumentError("invalid last_name", nil)
	}
	if user.Email != "" && (len(user.Email) > domain.MaxEmailLength || !strings.Contains(user.Email, "@")) {
		return domain.InvalidArgumentError("invalid email", nil)
	}
	if user.ToilerScore != 0 && (user.ToilerScore < domain.MinToilerScore || user.ToilerScore > domain.MaxToilerScore) {
		return domain.InvalidArgumentError("invalid toiler_score", nil)
	}
	if user.HeightCM != nil && (*user.HeightCM < 100 || *user.HeightCM > 260) {
		return domain.InvalidArgumentError("invalid height_cm", nil)
	}
	return nil
}

func ValidateUploadPhotoRequest(req domain.UploadPhotoRequest) error {
	if req.UserID == uuid.Nil {
		return domain.InvalidArgumentError("invalid user_id", nil)
	}
	if req.PhotoType != domain.PhotoTypePrimary && req.PhotoType != domain.PhotoTypeExtra {
		return domain.InvalidArgumentError("invalid photo_type", nil)
	}
	if req.ContentLength <= 0 || req.ContentLength > domain.MaxPhotoSizeBytes {
		return domain.InvalidArgumentError("invalid content_length", nil)
	}
	if strings.TrimSpace(req.ContentType) == "" || len(req.ContentType) > domain.MaxContentTypeLength {
		return domain.InvalidArgumentError("invalid content_type", nil)
	}
	switch req.ContentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return domain.InvalidArgumentError("unsupported content_type", nil)
	}
	if req.PhotoType == domain.PhotoTypeExtra {
		if req.ExtraPosition == nil || *req.ExtraPosition < domain.MinExtraPhotoPosition || *req.ExtraPosition > domain.MaxExtraPhotoPosition {
			return domain.InvalidArgumentError("invalid extra_position", nil)
		}
	}
	return nil
}
