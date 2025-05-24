package processor_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/http/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUntrackLinkDeleter_Handle(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("DeleteLink", context.Background(), int64(42), "https://example.com").
		Return(nil).
		Once()

	cache := mocks.NewMockCache(t)
	cache.On("InvalidateListLinks", mock.Anything, int64(42)).
		Return(nil).
		Once()

	handler := processor.NewUntrackLinkDeleter(client, channels, cache)

	state := &processor.State{
		Message:  "https://example.com",
		ChatID:   42,
		FSMState: "initial",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg message")

		expected := "Ссылка успешно удалена!"
		assert.Equal(t, expected, msg.Text)
	}()

	result := handler.Handle(context.Background(), state)
	assert.False(
		t,
		result.IsAutoTransition,
		"Expected auto transition is false",
	)
	assert.Equal(t, state, result.Result, "Expected result is nil")

	wg.Wait()
}

func TestUntrackLinkDeleter_Handle_UserError(t *testing.T) {
	t.Parallel()

	userErr := scrapper.ErrUserResponse{Message: "User error occurred"}
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("DeleteLink", context.Background(), int64(42), "https://example.com").
		Return(userErr).
		Once()

	cache := mocks.NewMockCache(t)

	handler := processor.NewUntrackLinkDeleter(client, channels, cache)

	state := &processor.State{
		Message:  "https://example.com",
		ChatID:   42,
		FSMState: "initial",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg message")

		expected := "User error occurred. Введите ссылку, которая отображается в /list"
		assert.Equal(t, expected, msg.Text, "Invalid message")
	}()

	result := handler.Handle(context.Background(), state)
	assert.False(
		t,
		result.IsAutoTransition,
		"Expected to be true",
	)
	assert.Equal(t, state, result.Result, "Expected to return the same state")

	wg.Wait()
}

func TestUntrackLinkDeleter_Handle_GenericError(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("generic error")
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("DeleteLink", context.Background(), int64(42), "https://example.com").
		Return(genericErr).
		Once()

	cache := mocks.NewMockCache(t)

	handler := processor.NewUntrackLinkDeleter(client, channels, cache)

	state := &processor.State{
		Message:  "https://example.com",
		ChatID:   42,
		FSMState: "initial",
	}

	result := handler.Handle(context.Background(), state)

	assert.Equal(t, "ошибка при удалении ссылки", state.ShowError)
	assert.Equal(
		t,
		"fail",
		result.NextState.String(),
		"Expected to be fail",
	)
	assert.True(
		t,
		result.IsAutoTransition,
		"Expected auto transition is true",
	)
	require.NotNil(t, result.Error, "Expected error")
}
