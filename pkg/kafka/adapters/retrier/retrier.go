package retrier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
	lastCheck       time.Time

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
		lastCheck:       time.Now(),
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
		gocron.NewTask(func() {
			if err := r.sendRetries(ctx); err != nil {
				slog.Error("failed to send retries", slog.Any("error", err))

				return
			}
		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	scheduler.Start()

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

			if msg.RetryCount() > r.cfg.MaxRetries {
				msg.NackToDLQ()

				continue
			}

			err := r.saveMessage(ctx, msg)
			if err != nil {
				slog.Error("failed to save message", slog.Any("error", err))

				continue
			}

			// mark as done
			msg.Ack()
		}
	}
}

func (r *Retrier) sendRetries(ctx context.Context) error {
	retryTime := time.Now()

	msgs, err := r.getRetryMessages(ctx, r.lastCheck, retryTime)
	if err != nil {
		return fmt.Errorf("failed to get retry messages: %w", err)
	}

	r.lastCheck = retryTime

	slog.Debug("trying to retry messages", slog.Int("count", len(msgs)))

	for _, msg := range msgs {
		r.channels.KafkaOutput() <- msg
	}

	return nil
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
