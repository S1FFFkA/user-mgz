package user

import (
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
)

func TestValidateUserForCreate_InvalidToilerScore(t *testing.T) {
	err := ValidateUserForCreate(domain.User{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		ToilerScore:           11,
		Sex:                   domain.SexMale,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoURL:       "u",
	})
	if err == nil || !domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument) {
		t.Fatalf("expected invalid_argument, got %v", err)
	}
}
