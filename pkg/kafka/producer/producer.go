package producer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/config"
)

type Closer func() error

type Channels interface {
	KafkaInput() chan *kafka.Input
}

type Producer struct {
	producer sarama.SyncProducer
	channels Channels

	closers []Closer
}

func New(cfg *config.Kafka, channels Channels) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = cfg.Producer.ReturnSuccesses
	config.Producer.Return.Errors = cfg.Producer.ReturnErrors
	config.Producer.RequiredAcks = sarama.RequiredAcks(cfg.Producer.RequiredAcks)
	config.Producer.Retry.Max = cfg.Producer.RetryMax

	switch cfg.Producer.Partitioner {
	case "random":
		config.Producer.Partitioner = sarama.NewRandomPartitioner

	case "roundrobin":
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	case "hash":
		config.Producer.Partitioner = sarama.NewHashPartitioner

	default:
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	closers := make([]Closer, 0)
	closers = append(closers, producer.Close)

	return &Producer{
		producer: producer,
		channels: channels,
		closers:  closers,
	}, nil
}

func (p *Producer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return p.close()

		case msg, ok := <-p.channels.KafkaInput():
			if !ok {
				slog.Warn("producer input channel closed")

				return p.close()
			}

			slog.Debug(
				"message to producer received",
				slog.Any("topic", msg.Topic),
				slog.Any("key", msg.Key),
				slog.Any("value", msg.Value),
			)

			err := p.produce(ctx, msg.Topic, msg.Key, msg.Value)
			if err != nil {
				slog.Error("failed to produce message", slog.Any("error", err))
			}
		}
	}
}

func (p *Producer) produce(ctx context.Context, topic, key, value string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (p *Producer) close() error {
	var err error

	for _, closer := range p.closers {
		if clErr := closer(); clErr != nil {
			err = errors.Join(err, clErr)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}

	return err
}
