package retrier

import (
	"context"
	"fmt"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	"github.com/georgysavva/scany/v2/pgxscan"
)

const createTableQuery = `
CREATE TABLE IF NOT EXISTS %s (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	key TEXT NOT NULL,
	value TEXT NOT NULL,
	topic TEXT NOT NULL,
	partition INT NOT NULL,
	kafka_offset INT NOT NULL,
	retry_count INT NOT NULL,
	retry_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);
`

func (r *Retrier) initTable(ctx context.Context) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(createTableQuery, r.cfg.TableName))
	if err != nil {
		return fmt.Errorf("failed to create %s table: %w", r.cfg.TableName, err)
	}

	return nil
}

const saveMessageQuery = `
INSERT INTO %s (key, value, topic, partition, kafka_offset, retry_count, retry_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

func (r *Retrier) saveMessage(ctx context.Context, msg *kafka.Message) error {
	dbMessage := r.kafkaMessageToDatabaseMessage(msg)

	_, err := r.db.Exec(
		ctx,
		fmt.Sprintf(saveMessageQuery, r.cfg.TableName),
		dbMessage.Key,
		dbMessage.Value,
		dbMessage.Topic,
		dbMessage.Partition,
		dbMessage.Offset,
		dbMessage.RetryCount,
		dbMessage.RetryAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

const getRetryMessagesQuery = `
SELECT id, key, value, topic, partition, kafka_offset, retry_count, retry_at, created_at
FROM %s
WHERE $1 < retry_at AND retry_at <= $2
`

func (r *Retrier) getRetryMessages(
	ctx context.Context,
	from, to time.Time,
) ([]*kafka.Message, error) {
	dbMessages := make([]*DatabaseMessage, 0)

	err := pgxscan.Select(
		ctx,
		r.db,
		&dbMessages,
		fmt.Sprintf(getRetryMessagesQuery, r.cfg.TableName),
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get retry messages: %w", err)
	}

	messages := make([]*kafka.Message, 0, len(dbMessages))

	for _, dbMessage := range dbMessages {
		messages = append(messages, r.databaseMessageToKafkaMessage(dbMessage))
	}

	return messages, nil
}
