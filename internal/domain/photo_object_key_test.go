package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestUserPhotoObjectKey_Primary(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	key := UserPhotoObjectKey(id, PhotoTypePrimary, nil, ".jpg")
	if key == "" {
		t.Fatalf("expected key")
	}
}

func TestUserPhotoObjectKey_Extra(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	pos := int16(2)
	key := UserPhotoObjectKey(id, PhotoTypeExtra, &pos, ".png")
	if key == "" {
		t.Fatalf("expected key")
	}
}

func TestPhotoFileExtensionFromContentType(t *testing.T) {
	if PhotoFileExtensionFromContentType("image/jpeg") != ".jpg" {
		t.Fatalf("expected .jpg")
	}
	if PhotoFileExtensionFromContentType("image/png") != ".png" {
		t.Fatalf("expected .png")
	}
}
