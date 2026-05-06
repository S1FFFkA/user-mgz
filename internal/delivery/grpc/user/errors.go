package user

import (
	"github.com/S1FFFkA/user-mgz/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

func (s *Server) handleError(method string, err error, fields ...zap.Field) error {
	appErr := domain.NormalizeAppError(err)
	st := status.Convert(domain.ToGRPCStatus(appErr))
	logFields := []zap.Field{
		zap.String("grpc_method", method),
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

	if domain.IsErrorCode(appErr, domain.ErrorCodeInternal) || domain.IsErrorCode(appErr, domain.ErrorCodeService) {
		s.logger.Error("request failed (5xx / internal)", append(logFields, zap.String("severity", "5xx"))...)
		return domain.ToGRPCStatus(appErr)
	}

	s.logger.Error("request failed", logFields...)
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
