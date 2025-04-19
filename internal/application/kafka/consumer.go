package kafka

import (
	"context"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

type Repository interface {
	AddUpdate(ctx context.Context, update *domain.Update) error
}

type Consumer struct {
	repo     Repository
	channels *domain.Channels
}

func NewConsumer(repo Repository, channels *domain.Channels) *Consumer {
	return &Consumer{
		repo:     repo,
		channels: channels,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case msg := <-c.channels.KafkaOutput():
			var update domain.Update

			if err := msg.Bind(&update); err != nil {
				slog.Error("failed to bind message", slog.Any("error", err))

				msg.Nack()

				continue
			}

			if err := c.repo.AddUpdate(ctx, &update); err != nil {
				slog.Error("failed to add update", slog.Any("error", err))

				msg.Nack()

				continue
			}

			msg.Ack()
		}
	}
}
