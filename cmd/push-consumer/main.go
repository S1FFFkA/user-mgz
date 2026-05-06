package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/S1FFFkA/user-mgz/internal/config"
	"github.com/S1FFFkA/user-mgz/internal/consumer/chatmessages"
	"github.com/S1FFFkA/user-mgz/internal/push"
	pushtokenrepo "github.com/S1FFFkA/user-mgz/internal/repository/pushtoken"
	pgstorage "github.com/S1FFFkA/user-mgz/internal/storage/postgres"
	"github.com/S1FFFkA/user-mgz/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.NewJSON()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()

	cfg, err := config.LoadPushConsumer()
	if err != nil {
		log.Fatal("failed to load push-consumer config", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgstorage.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to initialize postgres pool", zap.Error(err))
	}
	defer pool.Close()

	tokenRepo := pushtokenrepo.New(pool)
	provider, err := newPushProvider(ctx, cfg, log)
	if err != nil {
		log.Fatal("failed to init push provider", zap.Error(err))
	}

	consumer := chatmessages.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaConsumerGroup,
		cfg.KafkaTopicChatMessages,
		tokenRepo,
		provider,
		log,
	)

	log.Info("push consumer started",
		zap.Strings("kafka_brokers", cfg.KafkaBrokers),
		zap.String("kafka_topic", cfg.KafkaTopicChatMessages),
		zap.String("kafka_group", cfg.KafkaConsumerGroup),
		zap.String("provider", cfg.PushProvider),
	)
	if err = consumer.Run(ctx); err != nil {
		log.Fatal("push consumer stopped with error", zap.Error(err))
	}
	log.Info("push consumer stopped")
}

func newPushProvider(ctx context.Context, cfg config.PushConsumerConfig, log *zap.Logger) (push.Provider, error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.PushProvider))
	switch mode {
	case "noop":
		return push.NewNoopProvider(log), nil
	case "composite", "real", "live", "fcm", "firebase":
		return buildCompositeProvider(ctx, cfg, log)
	default:
		return nil, fmt.Errorf("unknown PUSH_PROVIDER %q (use noop or composite)", cfg.PushProvider)
	}
}

func buildCompositeProvider(ctx context.Context, cfg config.PushConsumerConfig, log *zap.Logger) (push.Provider, error) {
	credPath := cfg.FirebaseCredentialsPath
	if credPath == "" {
		credPath = strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	}
	jsonRaw := strings.TrimSpace(cfg.FirebaseCredentialsJSON)
	var jsonBytes []byte
	if jsonRaw != "" {
		jsonBytes = []byte(jsonRaw)
	}

	var fcm *push.FCMProvider
	if credPath != "" || len(jsonBytes) > 0 {
		var err error
		fcm, err = push.NewFCMProvider(ctx, credPath, jsonBytes, log)
		if err != nil {
			return nil, err
		}
	}

	var apns *push.APNSProvider
	hasAPNSKey := strings.TrimSpace(os.Getenv("APNS_KEY_PATH")) != "" ||
		strings.TrimSpace(os.Getenv("APNS_KEY_PEM")) != ""
	if hasAPNSKey {
		var err error
		apns, err = push.NewAPNSProviderFromEnv(log)
		if err != nil {
			return nil, err
		}
	}

	return push.NewCompositeProvider(fcm, apns, push.NewNoopProvider(log), log)
}
