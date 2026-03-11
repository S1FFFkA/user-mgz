package user

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	defaultUsersLimit     = 20
	maxUsersLimit         = 100
	defaultPresignTTL     = 15 * time.Minute
	defaultDownloadURLTTL = 15 * time.Minute
	maxFirstNameLength    = 80
	maxLastNameLength     = 80
	maxEmailLength        = 255
	maxContentTypeLength  = 100
)

type UseCase interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error)

	GetUserPhotoUploadURL(ctx context.Context, req domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error)
	ConfirmUserPhotoUpload(ctx context.Context, req domain.ConfirmPhotoUploadRequest) (domain.User, error)
	DeleteUserPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error
	GetUserPhotoDownloadURL(ctx context.Context, req domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error)
}

type Service struct {
	userRepo      repository.UserRepository
	userPhotoRepo repository.UserPhotoRepository
	s3Repo        repository.S3Repository
}

func NewService(
	userRepo repository.UserRepository,
	userPhotoRepo repository.UserPhotoRepository,
	s3Repo repository.S3Repository,
) *Service {
	return &Service{
		userRepo:      userRepo,
		userPhotoRepo: userPhotoRepo,
		s3Repo:        s3Repo,
	}
}

func (s *Service) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	if err := validateUserForCreate(user); err != nil {
		return domain.User{}, err
	}

	created, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, mapRepoError(err, "failed to create user")
	}

	if len(user.ExtraPhotos) > 0 {
		if err = s.userPhotoRepo.ReplaceExtraPhotos(ctx, created.ID, user.ExtraPhotos); err != nil {
			return domain.User{}, mapRepoError(err, "failed to save extra photos")
		}
	}

	created.ExtraPhotos, err = s.userPhotoRepo.ListExtraPhotos(ctx, created.ID)
	if err != nil {
		return domain.User{}, mapRepoError(err, "failed to fetch extra photos")
	}

	return created, nil
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	if userID == uuid.Nil {
		return domain.User{}, domain.InvalidArgumentError("invalid user_id", nil)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, mapRepoError(err, "user not found")
	}

	user.ExtraPhotos, err = s.userPhotoRepo.ListExtraPhotos(ctx, user.ID)
	if err != nil {
		return domain.User{}, mapRepoError(err, "failed to fetch extra photos")
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	if user.ID == uuid.Nil {
		return domain.User{}, domain.InvalidArgumentError("invalid user_id", nil)
	}
	if err := validateUserForUpdate(user); err != nil {
		return domain.User{}, err
	}

	updated, err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return domain.User{}, mapRepoError(err, "failed to update user")
	}

	if len(user.ExtraPhotos) > 0 {
		if err = s.userPhotoRepo.ReplaceExtraPhotos(ctx, user.ID, user.ExtraPhotos); err != nil {
			return domain.User{}, mapRepoError(err, "failed to replace extra photos")
		}
	}

	updated.ExtraPhotos, err = s.userPhotoRepo.ListExtraPhotos(ctx, user.ID)
	if err != nil {
		return domain.User{}, mapRepoError(err, "failed to fetch extra photos")
	}

	return updated, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return domain.InvalidArgumentError("invalid user_id", nil)
	}
	return mapRepoError(s.userRepo.DeleteUser(ctx, userID), "user not found")
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int32, cityID *int64) ([]domain.User, error) {
	if limit <= 0 {
		limit = defaultUsersLimit
	}
	if limit > maxUsersLimit {
		limit = maxUsersLimit
	}
	if offset < 0 {
		offset = 0
	}
	if cityID != nil && *cityID <= 0 {
		return nil, domain.InvalidArgumentError("invalid city_id", nil)
	}

	users, err := s.userRepo.ListUsers(ctx, limit, offset, cityID)
	if err != nil {
		return nil, mapRepoError(err, "failed to list users")
	}

	for i := range users {
		users[i].ExtraPhotos, err = s.userPhotoRepo.ListExtraPhotos(ctx, users[i].ID)
		if err != nil {
			return nil, mapRepoError(err, "failed to fetch extra photos")
		}
	}

	return users, nil
}

func (s *Service) GetUserPhotoUploadURL(ctx context.Context, req domain.UploadPhotoRequest) (domain.UploadPhotoTicket, error) {
	if err := validateUploadPhotoRequest(req); err != nil {
		return domain.UploadPhotoTicket{}, err
	}

	ext := extFromContentType(req.ContentType)
	objectKey := buildObjectKey(req.UserID, req.PhotoType, req.ExtraPosition, ext)

	uploadURL, err := s.s3Repo.PresignPutURL(ctx, objectKey, req.ContentType, req.ContentLength, defaultPresignTTL)
	if err != nil {
		return domain.UploadPhotoTicket{}, domain.ServiceError("failed to generate upload url", err)
	}

	return domain.UploadPhotoTicket{
		ObjectKey:        objectKey,
		UploadURL:        uploadURL,
		ExpiresInSeconds: int64(defaultPresignTTL.Seconds()),
	}, nil
}

func (s *Service) ConfirmUserPhotoUpload(ctx context.Context, req domain.ConfirmPhotoUploadRequest) (domain.User, error) {
	if req.UserID == uuid.Nil || strings.TrimSpace(req.ObjectKey) == "" {
		return domain.User{}, domain.InvalidArgumentError("invalid photo upload confirm request", nil)
	}
	if req.PhotoType != domain.PhotoTypePrimary && req.PhotoType != domain.PhotoTypeExtra {
		return domain.User{}, domain.InvalidArgumentError("invalid photo_type", nil)
	}

	if req.PhotoType == domain.PhotoTypePrimary {
		if err := s.userPhotoRepo.SetPrimaryPhoto(ctx, req.UserID, req.ObjectKey, req.ObjectKey); err != nil {
			return domain.User{}, mapRepoError(err, "failed to set primary photo")
		}
	} else {
		if req.ExtraPosition == nil || *req.ExtraPosition < domain.MinExtraPhotoPosition || *req.ExtraPosition > domain.MaxExtraPhotoPosition {
			return domain.User{}, domain.InvalidArgumentError("invalid extra_position", nil)
		}
		photo := domain.UserPhoto{
			UserID:    req.UserID,
			ObjectKey: req.ObjectKey,
			URL:       req.ObjectKey,
			Position:  *req.ExtraPosition,
		}
		if err := s.userPhotoRepo.UpsertExtraPhotoByPosition(ctx, req.UserID, photo); err != nil {
			return domain.User{}, mapRepoError(err, "failed to upsert extra photo")
		}
	}

	return s.GetUser(ctx, req.UserID)
}

func (s *Service) DeleteUserPhoto(ctx context.Context, userID uuid.UUID, photoID int64) error {
	_ = ctx
	_ = userID
	_ = photoID
	return domain.InvalidArgumentError("extra photo deletion is not allowed, use photo replacement", nil)
}

func (s *Service) GetUserPhotoDownloadURL(ctx context.Context, req domain.DownloadPhotoRequest) (domain.DownloadPhotoTicket, error) {
	if req.UserID == uuid.Nil {
		return domain.DownloadPhotoTicket{}, domain.InvalidArgumentError("invalid user_id", nil)
	}
	if req.PhotoType != domain.PhotoTypePrimary && req.PhotoType != domain.PhotoTypeExtra {
		return domain.DownloadPhotoTicket{}, domain.InvalidArgumentError("invalid photo_type", nil)
	}

	var objectKey string
	if req.PhotoType == domain.PhotoTypePrimary {
		user, err := s.userRepo.GetUserByID(ctx, req.UserID)
		if err != nil {
			return domain.DownloadPhotoTicket{}, mapRepoError(err, "user not found")
		}
		objectKey = user.PrimaryPhotoObjectKey
	} else {
		if req.PhotoID == nil || *req.PhotoID <= 0 {
			return domain.DownloadPhotoTicket{}, domain.InvalidArgumentError("invalid photo_id", nil)
		}
		photo, err := s.userPhotoRepo.GetExtraPhotoByID(ctx, req.UserID, *req.PhotoID)
		if err != nil {
			return domain.DownloadPhotoTicket{}, mapRepoError(err, "extra photo not found")
		}
		objectKey = photo.ObjectKey
	}

	fileName := filepath.Base(objectKey)
	downloadURL, err := s.s3Repo.PresignGetURL(ctx, objectKey, fileName, false, defaultDownloadURLTTL)
	if err != nil {
		return domain.DownloadPhotoTicket{}, domain.ServiceError("failed to generate download url", err)
	}

	return domain.DownloadPhotoTicket{
		DownloadURL:      downloadURL,
		ExpiresInSeconds: int64(defaultDownloadURLTTL.Seconds()),
	}, nil
}

func validateUserForCreate(user domain.User) error {
	if user.FirstName == "" || len(user.FirstName) > maxFirstNameLength {
		return domain.InvalidArgumentError("invalid first_name", nil)
	}
	if user.LastName == "" || len(user.LastName) > maxLastNameLength {
		return domain.InvalidArgumentError("invalid last_name", nil)
	}
	if user.Email == "" || len(user.Email) > maxEmailLength || !strings.Contains(user.Email, "@") {
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

func validateUserForUpdate(user domain.User) error {
	if user.FirstName != "" && len(user.FirstName) > maxFirstNameLength {
		return domain.InvalidArgumentError("invalid first_name", nil)
	}
	if user.LastName != "" && len(user.LastName) > maxLastNameLength {
		return domain.InvalidArgumentError("invalid last_name", nil)
	}
	if user.Email != "" && (len(user.Email) > maxEmailLength || !strings.Contains(user.Email, "@")) {
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

func validateUploadPhotoRequest(req domain.UploadPhotoRequest) error {
	if req.UserID == uuid.Nil {
		return domain.InvalidArgumentError("invalid user_id", nil)
	}
	if req.PhotoType != domain.PhotoTypePrimary && req.PhotoType != domain.PhotoTypeExtra {
		return domain.InvalidArgumentError("invalid photo_type", nil)
	}
	if req.ContentLength <= 0 || req.ContentLength > domain.MaxPhotoSizeBytes {
		return domain.InvalidArgumentError("invalid content_length", nil)
	}
	if strings.TrimSpace(req.ContentType) == "" || len(req.ContentType) > maxContentTypeLength {
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

func extFromContentType(contentType string) string {
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

func buildObjectKey(userID uuid.UUID, photoType domain.PhotoType, extraPosition *int16, ext string) string {
	if photoType == domain.PhotoTypePrimary {
		return fmt.Sprintf("users/%s/primary/%s%s", userID.String(), uuid.NewString(), ext)
	}
	if extraPosition == nil {
		return fmt.Sprintf("users/%s/extra/%s%s", userID.String(), uuid.NewString(), ext)
	}
	return fmt.Sprintf("users/%s/extra/%d/%s%s", userID.String(), *extraPosition, uuid.NewString(), ext)
}

func mapRepoError(err error, notFoundMessage string) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return domain.ConflictError("resource conflict", err)
		case "23503": // foreign_key_violation
			return domain.InvalidArgumentError("invalid reference", err)
		case "23514": // check_violation
			return domain.InvalidArgumentError("invalid request", err)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NotFoundError(notFoundMessage, err)
	}
	if domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument) ||
		domain.IsErrorCode(err, domain.ErrorCodeNotFound) ||
		domain.IsErrorCode(err, domain.ErrorCodeConflict) ||
		domain.IsErrorCode(err, domain.ErrorCodeUnauthorized) ||
		domain.IsErrorCode(err, domain.ErrorCodeForbidden) ||
		domain.IsErrorCode(err, domain.ErrorCodeService) ||
		domain.IsErrorCode(err, domain.ErrorCodeInternal) {
		return err
	}
	return domain.InternalError(err)
}
