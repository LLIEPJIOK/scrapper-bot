package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	"github.com/sony/gobreaker/v2"
)

type Producer struct {
	topic    string
	channels *domain.Channels
	cb       *gobreaker.CircuitBreaker[any]
}

func NewProducer(cfg *config.Kafka, channels *domain.Channels) *Producer {
	cbSettings := gobreaker.Settings{
		MaxRequests: cfg.CircuitBreaker.MaxHalfOpenRequests,
		Interval:    cfg.CircuitBreaker.Interval,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < cfg.CircuitBreaker.MinRequests {
				return false
			}

			return counts.ConsecutiveFailures >= cfg.CircuitBreaker.ConsecutiveFailures ||
				float64(
					counts.TotalFailures,
				)/float64(
					counts.Requests,
				) > cfg.CircuitBreaker.FailureRate
		},
	}
	cb := gobreaker.NewCircuitBreaker[any](cbSettings)

	return &Producer{
		topic:    cfg.UpdateTopic,
		channels: channels,
		cb:       cb,
	}
}

func (k *Producer) UpdatesPost(_ context.Context, update *domain.Update) error {
	raw, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	_, err = k.cb.Execute(func() (any, error) {
		err := kafka.Send(k.channels.KafkaInput(), &kafka.Input{
			Topic: k.topic,
			Value: string(raw),
			Key:   update.URL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send update: %w", err)
		}

		return "ok", nil
	})

	if err != nil {
		return client.NewErrServiceUnavailable(err)
	}

	return nil
}
