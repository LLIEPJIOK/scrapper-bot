package kafka_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/config"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/consumer"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/producer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testkafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type Channels struct {
	kafkaInput  chan *kafka.Input
	kafkaOutput chan *kafka.Message
}

func NewChannels() *Channels {
	return &Channels{
		kafkaInput:  make(chan *kafka.Input),
		kafkaOutput: make(chan *kafka.Message),
	}
}

func (c *Channels) KafkaInput() chan *kafka.Input {
	return c.kafkaInput
}

func (c *Channels) KafkaOutput() chan *kafka.Message {
	return c.kafkaOutput
}

func setupKafkaContainer(t *testing.T) (brokers []string, cleanup func()) {
	ctx := context.Background()

	kafkaContainer, err := testkafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		testkafka.WithClusterID("test-cluster"),
	)
	require.NoError(t, err, "failed to start kafka container")

	brokers, err = kafkaContainer.Brokers(ctx)
	require.NoError(t, err, "failed to get brokers")

	cleanup = func() {
		err := kafkaContainer.Terminate(ctx)
		assert.NoError(t, err, "failed to terminate kafka container")
	}

	return brokers, cleanup
}

func setupPostgresContainer(t *testing.T) (db *pgxpool.Pool, cleanup func()) {
	ctx := context.Background()

	var (
		dbName = "postgres"
		dbUser = "postgres"
		dbPass = "postgres"
	)

	postgresCont, err := testpostgres.Run(ctx,
		"postgres:latest",
		testpostgres.WithDatabase(dbName),
		testpostgres.WithUsername(dbUser),
		testpostgres.WithPassword(dbPass),
		testpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err, "failed to start postgres container")

	dsn, err := postgresCont.ConnectionString(ctx)
	require.NoError(t, err, "failed to get connection string")

	db, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err, "failed to open db")

	err = db.Ping(ctx)
	require.NoError(t, err, "failed to ping db")

	cleanup = func() {
		db.Close()

		err := postgresCont.Terminate(ctx)
		assert.NoError(t, err, "failed to terminate postgres container")
	}

	return db, cleanup
}

func TestKafka(t *testing.T) {
	brokers, kafkaCleanup := setupKafkaContainer(t)
	defer kafkaCleanup()

	db, dbCleanup := setupPostgresContainer(t)
	defer dbCleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Kafka{
		Brokers: brokers,
		Producer: config.Producer{
			RetryMax:        3,
			ReturnSuccesses: true,
			ReturnErrors:    true,
			RequiredAcks:    1,
			Partitioner:     "random",
		},
		Consumer: config.Consumer{
			Group: "test-group",
			Topics: []string{
				"test-topic",
			},
			ReturnErrors: true,
		},
		Retrier: config.Retrier{
			MaxRetries:    3,
			TableName:     "kafka_retrier",
			CheckInterval: 50 * time.Millisecond,
			InitialDelay:  100 * time.Millisecond,
		},
		DLQ: config.DLQ{
			TableName: "kafka_dlq",
		},
	}
	channels := NewChannels()

	kafkaProducer, err := producer.New(cfg, channels)
	require.NoError(t, err, "failed to create kafka producer")

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		err := kafkaProducer.Run(ctx)
		assert.NoError(t, err, "failed to run kafka producer")
	}()

	kafkaConsumer, err := consumer.New(cfg, db, channels)
	require.NoError(t, err, "failed to create kafka consumer")

	go func() {
		defer wg.Done()

		err := kafkaConsumer.Run(ctx)
		assert.NoError(t, err, "failed to run kafka consumer")
	}()

	// wait for the producer and consumer to start
	time.Sleep(time.Second)

	t.Run("success", func(t *testing.T) {
		err := kafka.Send(channels.KafkaInput(), &kafka.Input{
			Topic: "test-topic",
			Key:   "test-key",
			Value: "test-value",
		})
		require.NoError(t, err, "failed to send message to kafka")

		msg := <-channels.KafkaOutput()
		assert.Equal(t, "test-topic", msg.Base.Topic, "topic does not match")
		assert.Equal(t, "test-key", string(msg.Base.Key), "key does not match")
		assert.Equal(t, "test-value", string(msg.Base.Value), "value does not match")

		msg.Ack()

		// wait for saving data
		time.Sleep(time.Second)

		query := "SELECT COUNT(*) FROM kafka_retrier"

		var count int

		err = db.QueryRow(ctx, query).Scan(&count)
		require.NoError(t, err, "failed to get count")

		assert.Equal(t, 0, count, "count does not match")
	})

	t.Run("dlq", func(t *testing.T) {
		err := kafka.Send(channels.KafkaInput(), &kafka.Input{
			Topic: "test-topic",
			Key:   "test-key",
			Value: "test-value",
		})
		require.NoError(t, err, "failed to send message to kafka")

		for range cfg.Retrier.MaxRetries + 1 {
			msg := <-channels.KafkaOutput()
			assert.Equal(t, "test-topic", msg.Base.Topic, "topic does not match")
			assert.Equal(t, "test-key", string(msg.Base.Key), "key does not match")
			assert.Equal(t, "test-value", string(msg.Base.Value), "value does not match")

			msg.Nack()
		}

		retrierQuery := "SELECT COUNT(*) FROM kafka_retrier"

		var retrierCount int

		err = db.QueryRow(ctx, retrierQuery).Scan(&retrierCount)
		require.NoError(t, err, "failed to get count")

		// wait for saving data
		time.Sleep(time.Second)

		assert.Equal(t, 3, retrierCount, "count does not match")

		dlqQuery := "SELECT COUNT(*) FROM kafka_dlq"

		var dlqCount int

		err = db.QueryRow(ctx, dlqQuery).Scan(&dlqCount)
		require.NoError(t, err, "failed to get count")

		assert.Equal(t, 1, dlqCount, "count does not match")
	})

	cancel()
	wg.Wait()
}
