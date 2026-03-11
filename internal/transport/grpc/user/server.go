package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	grpcmw "github.com/S1FFFkA/user-mgz/internal/transport/grpc/middleware"
	usersvc "github.com/S1FFFkA/user-mgz/internal/usecase/user"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	userService usersvc.UseCase
	logger      *zap.Logger
}

func NewServer(userService usersvc.UseCase, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{
		userService: userService,
		logger:      logger,
	}
}

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user, err := fromCreateRequest(req)
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid create user request", err))
	}
	created, err := s.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, s.handleError(ctx, "CreateUser", err, zap.String("email", req.GetEmail()))
	}
	created = s.enrichUserPhotoLinks(ctx, created)
	return &userv1.CreateUserResponse{User: toProtoUser(created)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	user, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, s.handleError(ctx, "GetUser", err, zap.String("user_id", req.GetUserId()))
	}
	user = s.enrichUserPhotoLinks(ctx, user)
	return &userv1.GetUserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}

	current, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, s.handleError(ctx, "UpdateUser.GetUser", err, zap.String("user_id", req.GetUserId()))
	}

	updatedInput, err := mergeUpdateRequest(current, req)
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid update user request", err))
	}

	updated, err := s.userService.UpdateUser(ctx, updatedInput)
	if err != nil {
		return nil, s.handleError(ctx, "UpdateUser", err, zap.String("user_id", req.GetUserId()))
	}
	updated = s.enrichUserPhotoLinks(ctx, updated)
	return &userv1.UpdateUserResponse{User: toProtoUser(updated)}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	if err = s.userService.DeleteUser(ctx, userID); err != nil {
		return nil, s.handleError(ctx, "DeleteUser", err, zap.String("user_id", req.GetUserId()))
	}
	return &userv1.DeleteUserResponse{}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	var cityID *int64
	if req.GetCityId() > 0 {
		v := req.GetCityId()
		cityID = &v
	}
	users, err := s.userService.ListUsers(ctx, req.GetLimit(), req.GetOffset(), cityID)
	if err != nil {
		return nil, s.handleError(ctx, "ListUsers", err, zap.Int64("city_id", req.GetCityId()))
	}
	result := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		u = s.enrichUserPhotoLinks(ctx, u)
		result = append(result, toProtoUser(u))
	}
	return &userv1.ListUsersResponse{Users: result}, nil
}

func (s *Server) GetUserPhotoUploadUrl(ctx context.Context, req *userv1.GetUserPhotoUploadUrlRequest) (*userv1.GetUserPhotoUploadUrlResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	photoType, err := fromProtoPhotoType(req.GetPhotoType())
	if err != nil {
		return nil, domain.ToGRPCStatus(err)
	}
	var extraPos *int16
	if req.GetExtraPosition() > 0 {
		v := int16(req.GetExtraPosition())
		extraPos = &v
	}

	ticket, err := s.userService.GetUserPhotoUploadURL(ctx, domain.UploadPhotoRequest{
		UserID:        userID,
		PhotoType:     photoType,
		ExtraPosition: extraPos,
		ContentType:   req.GetContentType(),
		ContentLength: req.GetContentLength(),
	})
	if err != nil {
		return nil, s.handleError(ctx, "GetUserPhotoUploadUrl", err, zap.String("user_id", req.GetUserId()))
	}

	return &userv1.GetUserPhotoUploadUrlResponse{
		ObjectKey:        ticket.ObjectKey,
		UploadUrl:        ticket.UploadURL,
		ExpiresInSeconds: ticket.ExpiresInSeconds,
	}, nil
}

func (s *Server) ConfirmUserPhotoUpload(ctx context.Context, req *userv1.ConfirmUserPhotoUploadRequest) (*userv1.ConfirmUserPhotoUploadResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	photoType, err := fromProtoPhotoType(req.GetPhotoType())
	if err != nil {
		return nil, domain.ToGRPCStatus(err)
	}
	var extraPos *int16
	if req.GetExtraPosition() > 0 {
		v := int16(req.GetExtraPosition())
		extraPos = &v
	}
	user, err := s.userService.ConfirmUserPhotoUpload(ctx, domain.ConfirmPhotoUploadRequest{
		UserID:        userID,
		PhotoType:     photoType,
		ExtraPosition: extraPos,
		ObjectKey:     req.GetObjectKey(),
	})
	if err != nil {
		return nil, s.handleError(ctx, "ConfirmUserPhotoUpload", err, zap.String("user_id", req.GetUserId()))
	}
	user = s.enrichUserPhotoLinks(ctx, user)
	return &userv1.ConfirmUserPhotoUploadResponse{User: toProtoUser(user)}, nil
}

func (s *Server) DeleteUserPhoto(ctx context.Context, req *userv1.DeleteUserPhotoRequest) (*userv1.DeleteUserPhotoResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	if err = s.userService.DeleteUserPhoto(ctx, userID, req.GetPhotoId()); err != nil {
		return nil, s.handleError(ctx, "DeleteUserPhoto", err,
			zap.String("user_id", req.GetUserId()),
			zap.Int64("photo_id", req.GetPhotoId()),
		)
	}
	return &userv1.DeleteUserPhotoResponse{}, nil
}

func (s *Server) GetUserPhotoDownloadUrl(ctx context.Context, req *userv1.GetUserPhotoDownloadUrlRequest) (*userv1.GetUserPhotoDownloadUrlResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	photoType, err := fromProtoPhotoType(req.GetPhotoType())
	if err != nil {
		return nil, domain.ToGRPCStatus(err)
	}
	var photoID *int64
	if req.GetPhotoId() > 0 {
		v := req.GetPhotoId()
		photoID = &v
	}

	ticket, err := s.userService.GetUserPhotoDownloadURL(ctx, domain.DownloadPhotoRequest{
		UserID:    userID,
		PhotoType: photoType,
		PhotoID:   photoID,
	})
	if err != nil {
		return nil, s.handleError(ctx, "GetUserPhotoDownloadUrl", err, zap.String("user_id", req.GetUserId()))
	}

	return &userv1.GetUserPhotoDownloadUrlResponse{
		DownloadUrl:      ticket.DownloadURL,
		ExpiresInSeconds: ticket.ExpiresInSeconds,
	}, nil
}

func parseUUID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

func fromCreateRequest(req *userv1.CreateUserRequest) (domain.User, error) {
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

func mergeUpdateRequest(current domain.User, req *userv1.UpdateUserRequest) (domain.User, error) {
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

func fromProtoPhotoType(photoType userv1.PhotoType) (domain.PhotoType, error) {
	switch photoType {
	case userv1.PhotoType_PHOTO_TYPE_PRIMARY:
		return domain.PhotoTypePrimary, nil
	case userv1.PhotoType_PHOTO_TYPE_EXTRA:
		return domain.PhotoTypeExtra, nil
	default:
		return "", domain.InvalidArgumentError("invalid photo_type", errors.New(photoType.String()))
	}
}

func toProtoUser(user domain.User) *userv1.User {
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
		CreatedAt:             timestamppb.New(user.CreatedAt),
		UpdatedAt:             timestamppb.New(user.UpdatedAt),
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

func mapError(err error) error {
	if err == nil {
		return nil
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

func (s *Server) handleError(ctx context.Context, method string, err error, fields ...zap.Field) error {
	appErr := mapError(err)
	st := status.Convert(domain.ToGRPCStatus(appErr))
	logFields := []zap.Field{
		zap.String("grpc_method", method),
		zap.String("trace_id", grpcmw.TraceIDFromContext(ctx)),
		zap.String("grpc_code", st.Code().String()),
		zap.String("app_error_code", appErrorCode(appErr)),
		zap.Error(err),
	}
	logFields = append(logFields, fields...)

	if domain.IsErrorCode(appErr, domain.ErrorCodeInvalidArgument) ||
		domain.IsErrorCode(appErr, domain.ErrorCodeNotFound) ||
		domain.IsErrorCode(appErr, domain.ErrorCodeConflict) {
		s.logger.Warn("request failed validation", logFields...)
		return domain.ToGRPCStatus(appErr)
	}

	s.logger.Error("request failed with internal error", logFields...)
	return domain.ToGRPCStatus(appErr)
}

func appErrorCode(err error) string {
	switch {
	case domain.IsErrorCode(err, domain.ErrorCodeInvalidArgument):
		return string(domain.ErrorCodeInvalidArgument)
	case domain.IsErrorCode(err, domain.ErrorCodeNotFound):
		return string(domain.ErrorCodeNotFound)
	case domain.IsErrorCode(err, domain.ErrorCodeConflict):
		return string(domain.ErrorCodeConflict)
	case domain.IsErrorCode(err, domain.ErrorCodeUnauthorized):
		return string(domain.ErrorCodeUnauthorized)
	case domain.IsErrorCode(err, domain.ErrorCodeForbidden):
		return string(domain.ErrorCodeForbidden)
	case domain.IsErrorCode(err, domain.ErrorCodeService):
		return string(domain.ErrorCodeService)
	case domain.IsErrorCode(err, domain.ErrorCodeInternal):
		return string(domain.ErrorCodeInternal)
	default:
		return "unknown"
	}
}

func (s *Server) enrichUserPhotoLinks(ctx context.Context, user domain.User) domain.User {
	if user.ID == uuid.Nil {
		return user
	}

	primary, err := s.userService.GetUserPhotoDownloadURL(ctx, domain.DownloadPhotoRequest{
		UserID:    user.ID,
		PhotoType: domain.PhotoTypePrimary,
	})
	if err == nil && primary.DownloadURL != "" {
		user.PrimaryPhotoURL = primary.DownloadURL
	} else if err != nil {
		s.logger.Warn("failed to enrich primary photo download url",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}

	for i := range user.ExtraPhotos {
		photoID := user.ExtraPhotos[i].ID
		if photoID <= 0 {
			continue
		}
		ticket, e := s.userService.GetUserPhotoDownloadURL(ctx, domain.DownloadPhotoRequest{
			UserID:    user.ID,
			PhotoType: domain.PhotoTypeExtra,
			PhotoID:   &photoID,
		})
		if e == nil && ticket.DownloadURL != "" {
			user.ExtraPhotos[i].URL = ticket.DownloadURL
			continue
		}
		if e != nil {
			s.logger.Warn("failed to enrich extra photo download url",
				zap.String("user_id", user.ID.String()),
				zap.Int64("photo_id", photoID),
				zap.Error(e),
			)
		}
	}

	return user
}
