package domain

import (
	"time"

	"github.com/google/uuid"
)

type Sex string

const (
	SexMale   Sex = "male"
	SexFemale Sex = "female"
)

type PhotoType string

const (
	PhotoTypePrimary PhotoType = "primary"
	PhotoTypeExtra   PhotoType = "extra"
)

type City struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type UserPhoto struct {
	ID        int64
	UserID    uuid.UUID
	ObjectKey string
	URL       string
	Position  int16
	CreatedAt time.Time
}

type User struct {
	ID                    uuid.UUID
	FirstName             string
	LastName              string
	Email                 string
	BirthDate             time.Time
	Bio                   *string
	ToilerScore           int16
	AlcoholInfo           *string
	SmokingInfo           *string
	Sex                   Sex
	HeightCM              *int16
	CityID                *int64
	PrimaryPhotoObjectKey string
	PrimaryPhotoURL       string
	ExtraPhotos           []UserPhoto
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type UploadPhotoRequest struct {
	UserID        uuid.UUID
	PhotoType     PhotoType
	ExtraPosition *int16
	ContentType   string
	ContentLength int64
}

type UploadPhotoTicket struct {
	ObjectKey        string
	UploadURL        string
	ExpiresInSeconds int64
}

type ConfirmPhotoUploadRequest struct {
	UserID        uuid.UUID
	PhotoType     PhotoType
	ExtraPosition *int16
	ObjectKey     string
}

type DownloadPhotoRequest struct {
	UserID    uuid.UUID
	PhotoType PhotoType
	PhotoID   *int64
}

type DownloadPhotoTicket struct {
	DownloadURL      string
	ExpiresInSeconds int64
}
