package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// PhotoFileExtensionFromContentType возвращает расширение файла для ключа в S3.
func PhotoFileExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

// UserPhotoObjectKey строит object key для загрузки фото пользователя в S3.
func UserPhotoObjectKey(userID uuid.UUID, photoType PhotoType, extraPosition *int16, ext string) string {
	if photoType == PhotoTypePrimary {
		return fmt.Sprintf("users/%s/primary/%s%s", userID.String(), uuid.NewString(), ext)
	}
	if extraPosition == nil {
		return fmt.Sprintf("users/%s/extra/%s%s", userID.String(), uuid.NewString(), ext)
	}
	return fmt.Sprintf("users/%s/extra/%d/%s%s", userID.String(), *extraPosition, uuid.NewString(), ext)
}
