package user

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/delivery/userproto"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"go.uber.org/zap"
)

func (s *Server) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	var cityID *int64
	if req.GetCityId() > 0 {
		v := req.GetCityId()
		cityID = &v
	}
	users, err := s.users.ListUsers(ctx, req.GetLimit(), req.GetOffset(), cityID)
	if err != nil {
		return nil, s.handleError("ListUsers", err, zap.Int64("city_id", req.GetCityId()))
	}
	result := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		u = s.enrichUserPhotoLinks(ctx, u)
		result = append(result, userproto.ToProtoUser(u))
	}
	return &userv1.ListUsersResponse{Users: result}, nil
}
