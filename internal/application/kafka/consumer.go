package kafka

import (
	"context"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

			if update.SendImmediately.Value {
				tgMessage := tgbotapi.NewMessage(update.ChatID, update.Message)
				tgMessage.ParseMode = tgbotapi.ModeHTML

				c.channels.TelegramResp() <- tgMessage

				msg.Ack()

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
