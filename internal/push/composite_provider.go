package push

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// CompositeProvider отправляет уведомление в FCM или APNs по полю provider в БД.
type CompositeProvider struct {
	fcm      *FCMProvider
	apns     *APNSProvider
	fallback Provider
	log      *zap.Logger
}

func NewCompositeProvider(fcm *FCMProvider, apns *APNSProvider, fallback Provider, log *zap.Logger) (*CompositeProvider, error) {
	if fcm == nil && apns == nil {
		return nil, fmt.Errorf("push: хотя бы один из FCM или APNs должен быть сконфигурирован для реальной доставки")
	}
	if fallback == nil {
		fallback = &NoopProvider{logger: zap.NewNop()}
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &CompositeProvider{fcm: fcm, apns: apns, fallback: fallback, log: log}, nil
}

func (c *CompositeProvider) Send(ctx context.Context, n Notification) error {
	rowProvider := strings.TrimSpace(strings.ToLower(n.Provider))

	switch rowProvider {
	case "apns":
		if c.apns == nil {
			return fmt.Errorf("apns not configured but token has provider=apns")
		}
		return c.apns.Send(ctx, n)
	case "fcm", "firebase", "":
		if c.fcm != nil {
			return c.fcm.Send(ctx, n)
		}
		if c.apns != nil && strings.TrimSpace(strings.ToLower(n.Platform)) == "ios" {
			c.log.Warn("fcm not configured, trying apns for ios token", zap.String("recipient", n.RecipientUserID))
			return c.apns.Send(ctx, n)
		}
		return fmt.Errorf("fcm not configured (token provider=%q platform=%q)", n.Provider, n.Platform)
	default:
		if c.fcm != nil {
			return c.fcm.Send(ctx, n)
		}
		if c.apns != nil {
			return c.apns.Send(ctx, n)
		}
		return c.fallback.Send(ctx, n)
	}
}
