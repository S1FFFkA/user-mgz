// Package userproto — маппинг между protobuf API и доменными моделями (без зависимости от gRPC Server).
package userproto

import (
	"errors"
	"fmt"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"github.com/google/uuid"
)

func ParseUUID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

func FromCreateRequest(req *userv1.CreateUserRequest) (domain.User, error) {
	birthDate, err := time.Parse("2006-01-02", req.GetBirthDate())
	if err != nil {
		return domain.User{}, fmt.Errorf("parse birth_date: %w", err)
	}
	sex, err := fromProtoSex(req.GetSex())
	if err != nil {
		return domain.User{}, err
	}

	var bio *string
	if req.GetBio() != "" {
		v := req.GetBio()
		bio = &v
	}
	var alcohol *string
	if req.GetAlcoholInfo() != "" {
		v := req.GetAlcoholInfo()
		alcohol = &v
	}
	var smoking *string
	if req.GetSmokingInfo() != "" {
		v := req.GetSmokingInfo()
		smoking = &v
	}
	var heightCM *int16
	if req.GetHeightCm() > 0 {
		v := int16(req.GetHeightCm())
		heightCM = &v
	}
	var cityID *int64
	if req.GetCityId() > 0 {
		v := req.GetCityId()
		cityID = &v
	}

	extraPhotos := make([]domain.UserPhoto, 0, len(req.GetExtraPhotos()))
	for _, p := range req.GetExtraPhotos() {
		extraPhotos = append(extraPhotos, domain.UserPhoto{
			ObjectKey: p.GetObjectKey(),
			URL:       p.GetUrl(),
			Position:  int16(p.GetPosition()),
		})
	}

	return domain.User{
		FirstName:             req.GetFirstName(),
		LastName:              req.GetLastName(),
		Email:                 req.GetEmail(),
		BirthDate:             birthDate,
		Bio:                   bio,
		ToilerScore:           int16(req.GetToilerScore()),
		AlcoholInfo:           alcohol,
		SmokingInfo:           smoking,
		Sex:                   sex,
		HeightCM:              heightCM,
		CityID:                cityID,
		PrimaryPhotoObjectKey: req.GetPrimaryPhotoObjectKey(),
		PrimaryPhotoURL:       req.GetPrimaryPhotoUrl(),
		ExtraPhotos:           extraPhotos,
	}, nil
}

func MergeUpdateRequest(current domain.User, req *userv1.UpdateUserRequest) (domain.User, error) {
	out := current

	if req.FirstName != nil {
		out.FirstName = req.GetFirstName()
	}
	if req.LastName != nil {
		out.LastName = req.GetLastName()
	}
	if req.Email != nil {
		out.Email = req.GetEmail()
	}
	if req.BirthDate != nil {
		bd, err := time.Parse("2006-01-02", req.GetBirthDate())
		if err != nil {
			return domain.User{}, fmt.Errorf("parse birth_date: %w", err)
		}
		out.BirthDate = bd
	}
	if req.Bio != nil {
		v := req.GetBio()
		out.Bio = &v
	}
	if req.ToilerScore != nil {
		out.ToilerScore = int16(req.GetToilerScore())
	}
	if req.AlcoholInfo != nil {
		v := req.GetAlcoholInfo()
		out.AlcoholInfo = &v
	}
	if req.SmokingInfo != nil {
		v := req.GetSmokingInfo()
		out.SmokingInfo = &v
	}
	if req.Sex != nil {
		sex, err := fromProtoSex(req.GetSex())
		if err != nil {
			return domain.User{}, err
		}
		out.Sex = sex
	}
	if req.HeightCm != nil {
		v := int16(req.GetHeightCm())
		out.HeightCM = &v
	}
	if req.CityId != nil {
		v := req.GetCityId()
		out.CityID = &v
	}
	if req.PrimaryPhotoObjectKey != nil {
		out.PrimaryPhotoObjectKey = req.GetPrimaryPhotoObjectKey()
	}
	if req.PrimaryPhotoUrl != nil {
		out.PrimaryPhotoURL = req.GetPrimaryPhotoUrl()
	}
	if len(req.GetExtraPhotos()) > 0 {
		out.ExtraPhotos = out.ExtraPhotos[:0]
		for _, p := range req.GetExtraPhotos() {
			out.ExtraPhotos = append(out.ExtraPhotos, domain.UserPhoto{
				ObjectKey: p.GetObjectKey(),
				URL:       p.GetUrl(),
				Position:  int16(p.GetPosition()),
			})
		}
	}
	return out, nil
}

func fromProtoSex(sex userv1.Sex) (domain.Sex, error) {
	switch sex {
	case userv1.Sex_SEX_MALE:
		return domain.SexMale, nil
	case userv1.Sex_SEX_FEMALE:
		return domain.SexFemale, nil
	default:
		return "", domain.InvalidArgumentError("invalid sex", errors.New(sex.String()))
	}
}

func toProtoSex(sex domain.Sex) userv1.Sex {
	switch sex {
	case domain.SexMale:
		return userv1.Sex_SEX_MALE
	case domain.SexFemale:
		return userv1.Sex_SEX_FEMALE
	default:
		return userv1.Sex_SEX_UNSPECIFIED
	}
}

func FromProtoPhotoType(photoType userv1.PhotoType) (domain.PhotoType, error) {
	switch photoType {
	case userv1.PhotoType_PHOTO_TYPE_PRIMARY:
		return domain.PhotoTypePrimary, nil
	case userv1.PhotoType_PHOTO_TYPE_EXTRA:
		return domain.PhotoTypeExtra, nil
	default:
		return "", domain.InvalidArgumentError("invalid photo_type", errors.New(photoType.String()))
	}
}

func ToProtoUser(user domain.User) *userv1.User {
	result := &userv1.User{
		Id:                    user.ID.String(),
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		Email:                 user.Email,
		BirthDate:             user.BirthDate.Format("2006-01-02"),
		ToilerScore:           int32(user.ToilerScore),
		Sex:                   toProtoSex(user.Sex),
		PrimaryPhotoObjectKey: user.PrimaryPhotoObjectKey,
		PrimaryPhotoUrl:       user.PrimaryPhotoURL,
		CreatedAt:             timeToProtoTimestamp(user.CreatedAt),
		UpdatedAt:             timeToProtoTimestamp(user.UpdatedAt),
	}
	if user.Bio != nil {
		result.Bio = *user.Bio
	}
	if user.AlcoholInfo != nil {
		result.AlcoholInfo = *user.AlcoholInfo
	}
	if user.SmokingInfo != nil {
		result.SmokingInfo = *user.SmokingInfo
	}
	if user.HeightCM != nil {
		result.HeightCm = int32(*user.HeightCM)
	}
	if user.CityID != nil {
		result.City = &userv1.City{Id: *user.CityID}
	}
	result.ExtraPhotos = make([]*userv1.UserPhoto, 0, len(user.ExtraPhotos))
	for _, p := range user.ExtraPhotos {
		result.ExtraPhotos = append(result.ExtraPhotos, &userv1.UserPhoto{
			Id:        p.ID,
			ObjectKey: p.ObjectKey,
			Url:       p.URL,
			Position:  int32(p.Position),
		})
	}
	return result
}

func timeToProtoTimestamp(t time.Time) *userv1.Timestamp {
	if t.IsZero() {
		return nil
	}
	n := t.UnixNano()
	return &userv1.Timestamp{Seconds: n / 1e9, Nanos: int32(n % 1e9)}
}
