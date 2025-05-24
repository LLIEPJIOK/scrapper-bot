package dlq

import (
	"context"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Closer func() error

type DLQ struct {
	db  *pgxpool.Pool
	cfg *config.DLQ

	messageChannels *kafka.MessageChannels

	closers []Closer
}

func NewDLQ(
	cfg *config.DLQ,
	db *pgxpool.Pool,
	messageChannels *kafka.MessageChannels,
) *DLQ {
	return &DLQ{
		db:              db,
		cfg:             cfg,
		messageChannels: messageChannels,
		closers:         make([]Closer, 0),
	}
}

func (d *DLQ) Run(ctx context.Context) error {
	if err := d.initTable(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-d.messageChannels.DLQ():
			if !ok {
				slog.Warn("dlq channel closed")

				return nil
			}

			slog.Debug("message to dlq received")

			err := d.saveMessage(ctx, msg)
			if err != nil {
				slog.Error("failed to save message", slog.Any("error", err))

				continue
			}

			// mark as done
			msg.Ack()
		}
	}
}
