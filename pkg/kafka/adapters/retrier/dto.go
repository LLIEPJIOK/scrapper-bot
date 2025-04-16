package retrier

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
)

type DatabaseMessage struct {
	ID         int64     `db:"id"`
	Value      string    `db:"value"`
	Topic      string    `db:"topic"`
	Partition  int32     `db:"partition"`
	Offset     int64     `db:"offset"`
	RetryCount int32     `db:"retry_count"`
	RetryAt    time.Time `db:"retry_at"`
	CreatedAt  time.Time `db:"created_at"`
}

func (r *Retrier) databaseMessageToKafkaMessage(dbMessage *DatabaseMessage) *kafka.Message {
	base := &sarama.ConsumerMessage{
		Topic:     dbMessage.Topic,
		Value:     []byte(dbMessage.Value),
		Partition: dbMessage.Partition,
		Offset:    dbMessage.Offset,
	}

	return kafka.NewMessage(base, dbMessage.RetryCount, r.messageChannels)
}

func (r *Retrier) kafkaMessageToDatabaseMessage(kafkaMessage *kafka.Message) *DatabaseMessage {
	return &DatabaseMessage{
		Value:      string(kafkaMessage.Base.Value),
		Topic:      kafkaMessage.Base.Topic,
		Partition:  kafkaMessage.Base.Partition,
		Offset:     kafkaMessage.Base.Offset,
		RetryCount: kafkaMessage.RetryCount(),
		RetryAt:    time.Now().Add(r.cfg.InitialDelay * time.Duration(kafkaMessage.RetryCount())),
	}
}
