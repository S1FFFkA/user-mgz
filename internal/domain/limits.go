package domain

import "time"

// Ограничения полей пользователя (API / валидация).
const (
	MaxFirstNameLength   = 80
	MaxLastNameLength    = 80
	MaxEmailLength       = 255
	MaxContentTypeLength = 100
)

// Пагинация списка пользователей.
const (
	DefaultUsersListLimit int32 = 20
	MaxUsersListLimit     int32 = 100
)

// TTL presigned URL для S3 (по умолчанию; при необходимости вынести в config).
var (
	DefaultS3PutPresignTTL = 15 * time.Minute
	DefaultS3GetPresignTTL = 15 * time.Minute
)
