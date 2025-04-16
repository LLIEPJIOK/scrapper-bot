package retrier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/config"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Closer func() error

type Channels interface {
	KafkaOutput() chan *kafka.Message
}

type Retrier struct {
	db  *pgxpool.Pool
	cfg *config.Retrier

	channels        Channels
	messageChannels *kafka.MessageChannels

	closers []Closer
}

func NewRetrier(
	cfg *config.Retrier,
	db *pgxpool.Pool,
	channels Channels,
	messageChannels *kafka.MessageChannels,
) *Retrier {
	return &Retrier{
		db:              db,
		cfg:             cfg,
		channels:        channels,
		messageChannels: messageChannels,
		closers:         make([]Closer, 0),
	}
}

func (r *Retrier) Run(ctx context.Context) error {
	if err := r.initTable(ctx); err != nil {
		return err
	}

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(
			r.cfg.CheckInterval,
		),
		gocron.NewTask(r.sendRetries, ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	r.closers = append(r.closers, scheduler.Shutdown)

	for {
		select {
		case <-ctx.Done():
			return r.close()

		case msg, ok := <-r.messageChannels.Nack():
			if !ok {
				slog.Warn("retrier nack channel closed")

				return r.close()
			}

			slog.Debug("message to retrier received")

			if msg.RetryCount() >= r.cfg.MaxRetries {
				msg.Nack()

				continue
			}

			err := r.saveMessage(ctx, msg)
			if err != nil {
				slog.Error("failed to save message", slog.Any("error", err))
			}
		}
	}
}

func (r *Retrier) sendRetries(ctx context.Context) {
	msgs, err := r.getRetryMessages(ctx)
	if err != nil {
		slog.Error("failed to get retry messages", slog.Any("error", err))

		return
	}

	slog.Debug("trying to retry messages", slog.Int("count", len(msgs)))

	for _, msg := range msgs {
		r.channels.KafkaOutput() <- msg
	}
}

func (r *Retrier) close() error {
	var err error

	for _, closer := range r.closers {
		if clErr := closer(); clErr != nil {
			err = errors.Join(err, clErr)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to close retrier: %w", err)
	}

	return nil
}
