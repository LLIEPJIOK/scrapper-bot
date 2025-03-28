package processor_test

import (
	"context"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailer_Handle_NewMessage_NoError(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	failer := processor.NewFailer(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 0,
		ShowError: "",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg edit message")

		expectedText := "Не удалось обработать запрос. Попробуйте повторить запрос позже."

		assert.Equal(t, int64(123), msg.ChatID, "Expected ChatID to match")
		assert.Equal(t, expectedText, msg.Text, "Expected message text to match")
	}()

	result := failer.Handle(context.Background(), state)

	assert.False(t, result.IsAutoTransition, "Expected IsAutoTransition to be false")
	assert.Equal(t, state, result.Result, "Expected returned state to match input state")

	wg.Wait()
}

func TestFailer_Handle_NewMessage_WithError(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	failer := processor.NewFailer(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 0,
		ShowError: "Test error message",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg edit message")

		expectedText := "Не удалось обработать запрос: Test error message. Попробуйте повторить запрос позже."

		assert.Equal(t, int64(123), msg.ChatID, "Expected ChatID to match")
		assert.Equal(t, expectedText, msg.Text, "Expected message text to match")
	}()

	result := failer.Handle(context.Background(), state)

	assert.False(t, result.IsAutoTransition, "Expected IsAutoTransition to be false")
	assert.Equal(t, state, result.Result, "Expected returned state to match input state")

	wg.Wait()
}

func TestFailer_Handle_EditMessage(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	failer := processor.NewFailer(channels)

	state := &processor.State{
		ChatID:    456,
		MessageID: 789,
		ShowError: "",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.EditMessageTextConfig)
		require.True(t, ok, "not tg edit message")

		expectedText := "Не удалось обработать запрос. Попробуйте повторить запрос позже."

		assert.Equal(t, int64(456), msg.ChatID, "Expected ChatID to match")
		assert.Equal(t, 789, msg.MessageID, "Expected MessageID to match")
		assert.Equal(t, expectedText, msg.Text, "Expected message text to match")
	}()

	result := failer.Handle(context.Background(), state)

	assert.False(t, result.IsAutoTransition, "Expected IsAutoTransition to be false")
	assert.Equal(t, state, result.Result, "Expected returned state to match input state")

	wg.Wait()
}
