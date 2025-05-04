package kafka_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestProducer_UpdatesPost(t *testing.T) {
	testTopic := "test-topic"
	channels := domain.NewChannels()

	producer := kafka.NewProducer(testTopic, channels)

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
		assert.Equal(t, testTopic, msg.Topic, "topic does not match")
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
	testTopic := "test-topic"
	channels := domain.NewChannels()

	producer := kafka.NewProducer(testTopic, channels)

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
		assert.Equal(t, testTopic, msg.Topic, "topic does not match")
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
