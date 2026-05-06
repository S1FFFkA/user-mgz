package push

import "context"

type Notification struct {
	Provider        string
	Platform        string
	Token           string
	ChatID          string
	MessageID       int64
	SenderUserID    string
	RecipientUserID string
	MessageText     string
	UnreadCount     int64
}

type Provider interface {
	Send(ctx context.Context, n Notification) error
}
