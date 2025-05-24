package kafka_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/kafka/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	pkgkafka "github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConsumer_Run_ImmediateSending(t *testing.T) {
	channels := domain.NewChannels()
	messageChannels := pkgkafka.NewMessageChannels()

	update := &domain.Update{
		ChatID:          123456,
		Message:         "test message",
		SendImmediately: domain.NewNull(true),
	}

	raw, err := json.Marshal(update)
	assert.NoError(t, err, "failed to marshal update")

	mockRepo := mocks.NewMockRepository(t)
	consumer := kafka.NewConsumer(mockRepo, channels)

	saramaMessage := &sarama.ConsumerMessage{
		Value:     raw,
		Key:       []byte("key"),
		Partition: 0,
		Offset:    0,
	}
	msg := pkgkafka.NewMessage(saramaMessage, 0, messageChannels)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := consumer.Run(ctx)
		assert.NoError(t, err, "failed to run consumer")
	}()

	channels.KafkaOutput() <- msg

	resp := <-channels.TelegramResp()
	tgMessage, ok := resp.(tgbotapi.MessageConfig)
	assert.True(t, ok, "expected tgbotapi.MessageConfig")
	assert.Equal(t, update.Message, tgMessage.Text, "expected message to be sent to Telegram")

	msgAck := <-messageChannels.Ack()
	assert.Equal(t, msg, msgAck, "expected message to be acked")
}

func TestConsumer_Run_StoreInDB(t *testing.T) {
	channels := domain.NewChannels()

	update := &domain.Update{
		ChatID:          123456,
		Message:         "test message",
		SendImmediately: domain.NewNull(false),
	}

	raw, err := json.Marshal(update)
	assert.NoError(t, err, "failed to marshal update")

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.On("AddUpdate", mock.Anything, update).Return(nil).Once()

	consumer := kafka.NewConsumer(mockRepo, channels)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := consumer.Run(ctx)
		assert.NoError(t, err, "failed to run consumer")
	}()

	messageChannels := pkgkafka.NewMessageChannels()
	msg := pkgkafka.NewMessage(&sarama.ConsumerMessage{Value: raw}, 0, messageChannels)
	channels.KafkaOutput() <- msg

	select {
	case <-channels.TelegramResp():
		t.Error("expected message not to be sent to Telegram")
	case <-time.After(100 * time.Millisecond):
	}

	msgAck := <-messageChannels.Ack()
	assert.Equal(t, msg, msgAck, "expected message to be acked")
}

func TestConsumer_Run_StoreInDBError(t *testing.T) {
	channels := domain.NewChannels()

	update := &domain.Update{
		ChatID:          123456,
		Message:         "test message",
		SendImmediately: domain.NewNull(false),
	}

	raw, err := json.Marshal(update)
	assert.NoError(t, err, "failed to marshal update")

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.On("AddUpdate", mock.Anything, update).Return(assert.AnError).Once()

	consumer := kafka.NewConsumer(mockRepo, channels)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := consumer.Run(ctx)
		assert.NoError(t, err, "failed to run consumer")
	}()

	messageChannels := pkgkafka.NewMessageChannels()
	msg := pkgkafka.NewMessage(&sarama.ConsumerMessage{Value: raw}, 0, messageChannels)
	channels.KafkaOutput() <- msg

	select {
	case <-channels.TelegramResp():
		t.Error("expected message not to be sent to Telegram")
	case <-time.After(100 * time.Millisecond):
	}

	msgNack := <-messageChannels.Nack()
	assert.Equal(t, msg, msgNack, "expected message to be nacked")
}

func TestConsumer_Run_BindError(t *testing.T) {
	channels := domain.NewChannels()
	messageChannels := pkgkafka.NewMessageChannels()

	mockRepo := mocks.NewMockRepository(t)
	consumer := kafka.NewConsumer(mockRepo, channels)

	saramaMessage := &sarama.ConsumerMessage{
		Value:     []byte("invalid"),
		Key:       []byte("key"),
		Partition: 0,
		Offset:    0,
	}
	msg := pkgkafka.NewMessage(saramaMessage, 0, messageChannels)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := consumer.Run(ctx)
		assert.NoError(t, err, "failed to run consumer")
	}()

	channels.KafkaOutput() <- msg

	select {
	case <-channels.TelegramResp():
		t.Error("expected message not to be sent to Telegram")
	case <-time.After(100 * time.Millisecond):
	}

	msgNack := <-messageChannels.Nack()
	assert.Equal(t, msg, msgNack, "expected message to be nacked")
}
