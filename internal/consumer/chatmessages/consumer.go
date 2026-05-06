package chatmessages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/push"
	pushtokenrepo "github.com/S1FFFkA/user-mgz/internal/repository/pushtoken"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Event struct {
	Type            string    `json:"type"`
	ChatID          string    `json:"chat_id"`
	MessageID       int64     `json:"message_id"`
	UserID          string    `json:"user_id"`
	RecipientUserID string    `json:"recipient_user_id"`
	MessageText     string    `json:"message_text"`
	UnreadCount     int64     `json:"unread_count"`
	OccurredAt      time.Time `json:"occurred_at"`
}

type Consumer struct {
	reader       *kafka.Reader
	tokenRepo    *pushtokenrepo.Repository
	pushProvider push.Provider
	logger       *zap.Logger
}

func NewConsumer(brokers []string, groupID, topic string, tokenRepo *pushtokenrepo.Repository, provider push.Provider, logger *zap.Logger) *Consumer {
	if logger == nil {
		logger = zap.NewNop()
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	return &Consumer{
		reader:       reader,
		tokenRepo:    tokenRepo,
		pushProvider: provider,
		logger:       logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.reader.Close()
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("fetch kafka message: %w", err)
		}

		if handleErr := c.handleMessage(ctx, msg.Value); handleErr != nil {
			c.logger.Error("failed to handle kafka message", zap.Error(handleErr))
		}

		if err = c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit kafka message: %w", err)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, raw []byte) error {
	var event Event
	if err := json.Unmarshal(raw, &event); err != nil {
		return fmt.Errorf("unmarshal chat event: %w", err)
	}
	if event.Type != "message_created" {
		return nil
	}
	recipientID, err := uuid.Parse(event.RecipientUserID)
	if err != nil {
		return fmt.Errorf("parse recipient_user_id: %w", err)
	}

	tokens, err := c.tokenRepo.ListByUserID(ctx, recipientID)
	if err != nil {
		return fmt.Errorf("list push tokens: %w", err)
	}
	if len(tokens) == 0 {
		c.logger.Info("no push tokens for recipient", zap.String("recipient_user_id", event.RecipientUserID))
		return nil
	}

	for _, token := range tokens {
		if strings.TrimSpace(token.Token) == "" {
			continue
		}
		err = c.pushProvider.Send(ctx, push.Notification{
			Provider:        token.Provider,
			Platform:        token.Platform,
			Token:           token.Token,
			ChatID:          event.ChatID,
			MessageID:       event.MessageID,
			SenderUserID:    event.UserID,
			RecipientUserID: event.RecipientUserID,
			MessageText:     event.MessageText,
			UnreadCount:     event.UnreadCount,
		})
		if err != nil {
			c.logger.Error("push send failed",
				zap.Error(err),
				zap.String("recipient_user_id", event.RecipientUserID),
				zap.Int64("push_token_id", token.ID),
			)
		}
	}
	return nil
}
