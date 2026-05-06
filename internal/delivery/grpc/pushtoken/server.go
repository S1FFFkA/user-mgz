package pushtoken

import (
	"context"
	"strings"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	pushtokenrepo "github.com/S1FFFkA/user-mgz/internal/repository/pushtoken"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	userv1.UnimplementedPushTokenServiceServer
	repo *pushtokenrepo.Repository
	log  *zap.Logger
}

func NewServer(repo *pushtokenrepo.Repository, log *zap.Logger) *Server {
	if log == nil {
		log = zap.NewNop()
	}
	return &Server{repo: repo, log: log}
}

func (s *Server) RegisterPushToken(ctx context.Context, req *userv1.RegisterPushTokenRequest) (*userv1.RegisterPushTokenResponse, error) {
	userID, err := uuid.Parse(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	token := strings.TrimSpace(req.GetToken())
	deviceID := strings.TrimSpace(req.GetDeviceId())
	platform := strings.TrimSpace(strings.ToLower(req.GetPlatform()))
	provider := strings.TrimSpace(strings.ToLower(req.GetProvider()))
	if token == "" {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("token is required", nil))
	}
	if deviceID == "" {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("device_id is required", nil))
	}
	if platform == "" {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("platform is required", nil))
	}
	if provider == "" {
		provider = "fcm"
	}

	if err := s.repo.Upsert(ctx, userID, provider, token, deviceID, platform); err != nil {
		s.log.Error("RegisterPushToken failed", zap.Error(err), zap.String("user_id", req.GetUserId()))
		return nil, status.Error(codes.Internal, "failed to save push token")
	}
	return &userv1.RegisterPushTokenResponse{}, nil
}

func (s *Server) RemovePushToken(ctx context.Context, req *userv1.RemovePushTokenRequest) (*userv1.RemovePushTokenResponse, error) {
	userID, err := uuid.Parse(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("invalid user_id", err))
	}
	deviceID := strings.TrimSpace(req.GetDeviceId())
	if deviceID == "" {
		return nil, domain.ToGRPCStatus(domain.InvalidArgumentError("device_id is required", nil))
	}
	if err := s.repo.DeleteByUserAndDevice(ctx, userID, deviceID); err != nil {
		s.log.Error("RemovePushToken failed", zap.Error(err), zap.String("user_id", req.GetUserId()))
		return nil, status.Error(codes.Internal, "failed to remove push token")
	}
	return &userv1.RemovePushTokenResponse{}, nil
}
