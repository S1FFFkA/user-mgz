package config

import (
	"errors"
	"os"
	"strings"
)

type PushConsumerConfig struct {
	DatabaseURL             string
	KafkaBrokers            []string
	KafkaTopicChatMessages  string
	KafkaConsumerGroup      string
	PushProvider            string
	FirebaseCredentialsPath string
	FirebaseCredentialsJSON string
}

func LoadPushConsumer() (PushConsumerConfig, error) {
	kafkaCSV := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	brokers := parseCSV(kafkaCSV)
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	cfg := PushConsumerConfig{
		DatabaseURL:             getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/user_service?sslmode=disable"),
		KafkaBrokers:            brokers,
		KafkaTopicChatMessages:  getEnvOrDefault("KAFKA_TOPIC_CHAT_MESSAGES", "chat.messages"),
		KafkaConsumerGroup:      getEnvOrDefault("KAFKA_CONSUMER_GROUP", "user-mgz-push-consumer"),
		PushProvider:            getEnvOrDefault("PUSH_PROVIDER", "noop"),
		FirebaseCredentialsPath: strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_PATH")),
		FirebaseCredentialsJSON: os.Getenv("FIREBASE_CREDENTIALS_JSON"),
	}
	if cfg.DatabaseURL == "" {
		return PushConsumerConfig{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
