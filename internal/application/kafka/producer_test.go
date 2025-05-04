package kafka_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(t *testing.T) *config.Kafka {
	t.Helper()

	return &config.Kafka{
		UpdateTopic: "test-topic",
		CircuitBreaker: config.CircuitBreaker{
			MaxHalfOpenRequests: 1,
			Interval:            50 * time.Millisecond,
			Timeout:             100 * time.Millisecond,
			MinRequests:         4,
			ConsecutiveFailures: 4,
			FailureRate:         0.6,
		},
	}
}

func TestProducer_UpdatesPost(t *testing.T) {
	cfg := newTestConfig(t)
	channels := domain.NewChannels()

	producer := kafka.NewProducer(cfg, channels)

	testUpdate := &domain.Update{
		URL:             "test-url",
		Message:         "test-message",
		ChatID:          123456,
		SendImmediately: domain.NewNull(true),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		msg := <-channels.KafkaInput()
		assert.Equal(t, cfg.UpdateTopic, msg.Topic, "topic does not match")
		assert.Equal(t, testUpdate.URL, msg.Key, "key does not match")

		var receivedUpdate domain.Update
		err := json.Unmarshal([]byte(msg.Value), &receivedUpdate)
		assert.NoError(t, err, "failed to unmarshal update")
		assert.Equal(t, testUpdate, &receivedUpdate, "update does not match")

		msg.ResChan <- nil
	}()

	err := producer.UpdatesPost(context.Background(), testUpdate)
	assert.NoError(t, err, "failed to send update")

	wg.Wait()
}

func TestProducer_UpdatesPost_Error(t *testing.T) {
	cfg := newTestConfig(t)
	channels := domain.NewChannels()

	producer := kafka.NewProducer(cfg, channels)

	testUpdate := &domain.Update{
		URL:             "test-url",
		Message:         "test-message",
		ChatID:          123456,
		SendImmediately: domain.NewNull(true),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		msg := <-channels.KafkaInput()
		assert.Equal(t, cfg.UpdateTopic, msg.Topic, "topic does not match")
		assert.Equal(t, testUpdate.URL, msg.Key, "key does not match")

		var receivedUpdate domain.Update
		err := json.Unmarshal([]byte(msg.Value), &receivedUpdate)
		assert.NoError(t, err, "failed to unmarshal update")
		assert.Equal(t, testUpdate, &receivedUpdate, "update does not match")

		msg.ResChan <- assert.AnError
	}()

	err := producer.UpdatesPost(context.Background(), testUpdate)
	assert.Error(t, err, "should fail")

	wg.Wait()
}

func TestProducer_UpdatesPost_CircuitBreaker(t *testing.T) {
	cfg := newTestConfig(t)
	channels := domain.NewChannels()

	producer := kafka.NewProducer(cfg, channels)

	testUpdate := &domain.Update{
		URL:             "test-url",
		Message:         "test-message",
		ChatID:          123456,
		SendImmediately: domain.NewNull(true),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for range cfg.CircuitBreaker.ConsecutiveFailures {
			msg := <-channels.KafkaInput()
			assert.Equal(t, cfg.UpdateTopic, msg.Topic, "topic does not match")
			assert.Equal(t, testUpdate.URL, msg.Key, "key does not match")

			var receivedUpdate domain.Update
			err := json.Unmarshal([]byte(msg.Value), &receivedUpdate)
			assert.NoError(t, err, "failed to unmarshal update")
			assert.Equal(t, testUpdate, &receivedUpdate, "update does not match")

			msg.ResChan <- assert.AnError
		}

		msg := <-channels.KafkaInput()
		assert.Equal(t, cfg.UpdateTopic, msg.Topic, "topic does not match")
		assert.Equal(t, testUpdate.URL, msg.Key, "key does not match")

		var receivedUpdate domain.Update
		err := json.Unmarshal([]byte(msg.Value), &receivedUpdate)
		assert.NoError(t, err, "failed to unmarshal update")
		assert.Equal(t, testUpdate, &receivedUpdate, "update does not match")

		msg.ResChan <- nil
	}()

	for range cfg.CircuitBreaker.ConsecutiveFailures {
		err := producer.UpdatesPost(context.Background(), testUpdate)
		require.Error(t, err, "should fail")
		assert.ErrorIs(t, err, assert.AnError, "should be a test error")
	}

	err := producer.UpdatesPost(context.Background(), testUpdate)
	require.Error(t, err, "should fail")
	assert.ErrorIs(t, err, gobreaker.ErrOpenState, "cb should be open")

	// simulate recovery
	time.Sleep(2 * cfg.CircuitBreaker.Timeout)

	err = producer.UpdatesPost(context.Background(), testUpdate)
	require.NoError(t, err, "failed to update post")

	wg.Wait()
}
