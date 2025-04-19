package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/adapters/dlq"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/adapters/retrier"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Closer func() error

type Channels interface {
	KafkaOutput() chan *kafka.Message
}

type Consumer struct {
	consumer sarama.ConsumerGroup
	topics   []string
	closers  []Closer

	retrier *retrier.Retrier
	dlq     *dlq.DLQ

	channels        Channels
	messageChannels *kafka.MessageChannels
}

func New(
	cfg *config.Kafka,
	db *pgxpool.Pool,
	channels Channels,
) (*Consumer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Return.Errors = cfg.Consumer.ReturnErrors

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.Consumer.Group, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	closers := make([]Closer, 0)
	closers = append(closers, consumer.Close)

	messageChannels := kafka.NewMessageChannels()

	return &Consumer{
		consumer:        consumer,
		topics:          cfg.Consumer.Topics,
		closers:         closers,
		retrier:         retrier.NewRetrier(&cfg.Retrier, db, channels, messageChannels),
		dlq:             dlq.NewDLQ(&cfg.DLQ, db, messageChannels),
		channels:        channels,
		messageChannels: messageChannels,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return c.retrier.Run(ctx)
	})

	eg.Go(func() error {
		return c.dlq.Run(ctx)
	})

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return c.close()

			case err, ok := <-c.consumer.Errors():
				if !ok {
					slog.Warn("consumer error channel closed")

					return c.close()
				}

				slog.Error("consumer error", slog.Any("error", err))

			default:
				h := NewConsumerGroupHandler(c.channels, c.messageChannels)

				err := c.consumer.Consume(ctx, c.topics, h)
				if err != nil {
					slog.Error("consumer error", slog.Any("error", err))
				}
			}
		}
	})

	return eg.Wait()
}

func (c *Consumer) close() error {
	var err error

	for _, closer := range c.closers {
		if clErr := closer(); clErr != nil {
			err = errors.Join(err, clErr)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}

	return nil
}
