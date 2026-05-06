package push

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sideshow/apns2"
	apnspayload "github.com/sideshow/apns2/payload"
	apnstoken "github.com/sideshow/apns2/token"
	"go.uber.org/zap"
)

type APNSProvider struct {
	client *apns2.Client
	topic  string
	log    *zap.Logger
}

type apnsJWTConfig struct {
	KeyPEMBytes []byte
	KeyID       string
	TeamID      string
	Topic       string
	Sandbox     bool
}

func NewAPNSProviderFromEnv(log *zap.Logger) (*APNSProvider, error) {
	raw := strings.TrimSpace(os.Getenv("APNS_KEY_PEM"))
	path := strings.TrimSpace(os.Getenv("APNS_KEY_PATH"))

	var keyBytes []byte
	switch {
	case raw != "":
		keyBytes = []byte(raw)
	case path != "":
		var err error
		keyBytes, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read APNS key file: %w", err)
		}
	default:
		return nil, fmt.Errorf("APNs: set APNS_KEY_PATH or APNS_KEY_PEM")
	}

	cfg := apnsJWTConfig{
		KeyPEMBytes: keyBytes,
		KeyID:       strings.TrimSpace(os.Getenv("APNS_KEY_ID")),
		TeamID:      strings.TrimSpace(os.Getenv("APNS_TEAM_ID")),
		Topic:       strings.TrimSpace(os.Getenv("APNS_TOPIC")),
	}
	cfg.Sandbox = false
	if raw := strings.TrimSpace(os.Getenv("APNS_SANDBOX")); raw != "" {
		sb, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("APNS_SANDBOX must be true or false")
		}
		cfg.Sandbox = sb
	}

	if cfg.KeyID == "" || cfg.TeamID == "" || cfg.Topic == "" {
		return nil, fmt.Errorf("APNs: APNS_KEY_ID, APNS_TEAM_ID and APNS_TOPIC are required")
	}

	authKey, err := apnstoken.AuthKeyFromBytes(cfg.KeyPEMBytes)
	if err != nil {
		return nil, fmt.Errorf("parse APNs p8 key: %w", err)
	}

	token := &apnstoken.Token{
		AuthKey: authKey,
		KeyID:   cfg.KeyID,
		TeamID:  cfg.TeamID,
	}

	var client *apns2.Client
	if cfg.Sandbox {
		client = apns2.NewTokenClient(token).Development()
	} else {
		client = apns2.NewTokenClient(token).Production()
	}

	if log == nil {
		log = zap.NewNop()
	}
	return &APNSProvider{client: client, topic: cfg.Topic, log: log}, nil
}

func (p *APNSProvider) Send(ctx context.Context, n Notification) error {
	_ = ctx

	title := "Новое сообщение"
	body := n.MessageText
	if len(body) > 280 {
		body = body[:280] + "…"
	}

	pl := apnspayload.NewPayload().
		AlertTitle(title).
		AlertBody(body).
		Badge(int(n.UnreadCount)).
		ThreadID(n.ChatID).
		Sound("default")

	custom := map[string]interface{}{
		"chat_id":           n.ChatID,
		"message_id":        n.MessageID,
		"sender_user_id":    n.SenderUserID,
		"recipient_user_id": n.RecipientUserID,
		"unread_count":      n.UnreadCount,
	}
	for k, v := range custom {
		pl.Custom(k, v)
	}

	payloadBytes, err := pl.MarshalJSON()
	if err != nil {
		return fmt.Errorf("apns marshal payload: %w", err)
	}

	notif := &apns2.Notification{
		DeviceToken: strings.ReplaceAll(strings.TrimSpace(n.Token), " ", ""),
		Topic:       p.topic,
		Payload:     payloadBytes,
		Priority:    apns2.PriorityHigh,
		PushType:    apns2.PushTypeAlert,
	}

	resp, err := p.client.PushWithContext(ctx, notif)
	if err != nil {
		return fmt.Errorf("apns push: %w", err)
	}
	if !resp.Sent() {
		return fmt.Errorf("apns not sent: status=%d reason=%q apns-id=%q", resp.StatusCode, resp.Reason, resp.ApnsID)
	}

	p.log.Debug("apns delivered",
		zap.String("chat_id", n.ChatID),
		zap.Int64("message_id", n.MessageID),
		zap.String("recipient_user_id", n.RecipientUserID),
	)
	return nil
}
