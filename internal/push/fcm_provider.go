package push

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type FCMProvider struct {
	client *messaging.Client
	log    *zap.Logger
}

func NewFCMProvider(ctx context.Context, credentialsPath string, credentialsJSON []byte, log *zap.Logger) (*FCMProvider, error) {
	if credentialsPath == "" && len(credentialsJSON) == 0 {
		return nil, fmt.Errorf("FCM: set FIREBASE_CREDENTIALS_JSON or FIREBASE_CREDENTIALS_PATH (or GOOGLE_APPLICATION_CREDENTIALS)")
	}

	var opts []option.ClientOption
	switch {
	case len(credentialsJSON) > 0:
		opts = append(opts, option.WithCredentialsJSON(credentialsJSON))
	default:
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase app: %w", err)
	}
	cli, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging: %w", err)
	}

	if log == nil {
		log = zap.NewNop()
	}
	return &FCMProvider{client: cli, log: log}, nil
}

func (p *FCMProvider) Send(ctx context.Context, n Notification) error {
	title := "Новое сообщение"
	body := n.MessageText
	if len(body) > 280 {
		body = body[:280] + "…"
	}

	data := map[string]string{
		"chat_id":           n.ChatID,
		"message_id":        fmt.Sprintf("%d", n.MessageID),
		"sender_user_id":    n.SenderUserID,
		"recipient_user_id": n.RecipientUserID,
		"unread_count":      fmt.Sprintf("%d", n.UnreadCount),
	}

	msg := &messaging.Message{
		Token: n.Token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{"apns-priority": "10"},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Badge:    badgeFromUnread(n.UnreadCount),
					Sound:    "default",
					ThreadID: n.ChatID,
				},
			},
		},
	}

	_, err := p.client.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("fcm send: %w", err)
	}

	p.log.Debug("fcm delivered",
		zap.String("chat_id", n.ChatID),
		zap.Int64("message_id", n.MessageID),
		zap.String("recipient_user_id", n.RecipientUserID),
	)
	return nil
}

func badgeFromUnread(n int64) *int {
	if n < 0 {
		n = 0
	}
	const maxBadge = 999_999
	if n > maxBadge {
		n = maxBadge
	}
	b := int(n)
	return &b
}
