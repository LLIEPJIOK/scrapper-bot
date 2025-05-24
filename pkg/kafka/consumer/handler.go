package consumer

import (
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
)

type GroupHandler struct {
	channels        Channels
	messageChannels *kafka.MessageChannels
}

func NewConsumerGroupHandler(
	channels Channels,
	messageChannels *kafka.MessageChannels,
) *GroupHandler {
	return &GroupHandler{
		channels:        channels,
		messageChannels: messageChannels,
	}
}

func (GroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (GroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h GroupHandler) ConsumeClaim(
	sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	go h.ackMessages(sess)

	for msg := range claim.Messages() {
		slog.Debug(
			"message received",
			slog.Any("topic", msg.Topic),
			slog.Any("value", string(msg.Value)),
			slog.Any("offset", msg.Offset),
			slog.Any("partition", msg.Partition),
		)

		h.channels.KafkaOutput() <- kafka.NewMessage(msg, 0, h.messageChannels)
	}

	return nil
}

func (h GroupHandler) ackMessages(
	sess sarama.ConsumerGroupSession,
) {
	for {
		select {
		case <-sess.Context().Done():
			return

		case msg := <-h.messageChannels.Ack():
			slog.Debug(
				"message acked",
				slog.Any("topic", msg.Base.Topic),
				slog.Any("value", string(msg.Base.Value)),
				slog.Any("offset", msg.Base.Offset),
				slog.Any("partition", msg.Base.Partition),
			)

			sess.MarkMessage(msg.Base, "")
		}
	}
}
