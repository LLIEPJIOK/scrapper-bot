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
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = cfg.Producer.ReturnSuccesses
	saramaConfig.Producer.Return.Errors = cfg.Producer.ReturnErrors
	saramaConfig.Producer.RequiredAcks = sarama.RequiredAcks(cfg.Producer.RequiredAcks)
	saramaConfig.Producer.Retry.Max = cfg.Producer.RetryMax

	switch cfg.Producer.Partitioner {
	case "random":
		saramaConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	case "roundrobin":
		saramaConfig.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	case "hash":
		saramaConfig.Producer.Partitioner = sarama.NewHashPartitioner

	default:
		saramaConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
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
			msg.ResChan <- err
		}
	}
}

func (p *Producer) produce(_ context.Context, topic, key, value string) error {
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
