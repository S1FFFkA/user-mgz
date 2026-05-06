package push

import (
	"context"

	"go.uber.org/zap"
)

type NoopProvider struct {
	logger *zap.Logger
}

func NewNoopProvider(logger *zap.Logger) *NoopProvider {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &NoopProvider{logger: logger}
}

func (p *NoopProvider) Send(_ context.Context, n Notification) error {
	p.logger.Info("push noop send",
		zap.String("provider", n.Provider),
		zap.String("token", n.Token),
		zap.String("chat_id", n.ChatID),
		zap.Int64("message_id", n.MessageID),
		zap.String("sender_user_id", n.SenderUserID),
		zap.String("recipient_user_id", n.RecipientUserID),
		zap.String("message_text", n.MessageText),
		zap.Int64("unread_count", n.UnreadCount),
	)
	return nil
}
