package userproto

import (
	"testing"
	"time"

	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
)

func TestFromCreateRequest_ParsesBirthDate(t *testing.T) {
	_, err := FromCreateRequest(&userv1.CreateUserRequest{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             "not-a-date",
		ToilerScore:           7,
		Sex:                   userv1.Sex_SEX_MALE,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoUrl:       "u",
	})
	if err == nil {
		t.Fatalf("expected error")
	}

	u, err := FromCreateRequest(&userv1.CreateUserRequest{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             "2000-01-02",
		ToilerScore:           7,
		Sex:                   userv1.Sex_SEX_MALE,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoUrl:       "u",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.BirthDate != time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected birthdate: %v", u.BirthDate)
	}
}
