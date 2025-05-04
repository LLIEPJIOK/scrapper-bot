package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
)

type Producer struct {
	topic    string
	channels *domain.Channels
}

func NewProducer(topic string, channels *domain.Channels) *Producer {
	return &Producer{
		topic:    topic,
		channels: channels,
	}
}

func (k *Producer) UpdatesPost(_ context.Context, update *domain.Update) error {
	raw, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	err = kafka.Send(k.channels.KafkaInput(), &kafka.Input{
		Topic: k.topic,
		Value: string(raw),
		Key:   update.URL,
	})
	if err != nil {
		return fmt.Errorf("failed to send update: %w", err)
	}

	return nil
}
