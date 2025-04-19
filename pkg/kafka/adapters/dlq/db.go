package dlq

import (
	"context"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
)

const createTableQuery = `
CREATE TABLE IF NOT EXISTS %s (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	value TEXT NOT NULL,
	topic TEXT NOT NULL,
	partition INT NOT NULL,
	kafka_offset INT NOT NULL,
	retry_count INT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);
`

func (d *DLQ) initTable(ctx context.Context) error {
	_, err := d.db.Exec(ctx, fmt.Sprintf(createTableQuery, d.cfg.TableName))
	if err != nil {
		return fmt.Errorf("failed to create %s table: %w", d.cfg.TableName, err)
	}

	return nil
}

const saveMessageQuery = `
INSERT INTO %s (value, topic, partition, kafka_offset, retry_count)
VALUES ($1, $2, $3, $4, $5)
`

func (d *DLQ) saveMessage(ctx context.Context, msg *kafka.Message) error {
	_, err := d.db.Exec(
		ctx,
		fmt.Sprintf(saveMessageQuery, d.cfg.TableName),
		string(msg.Base.Value),
		string(msg.Base.Key),
		msg.Base.Partition,
		msg.Base.Offset,
		msg.RetryCount(),
	)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}
