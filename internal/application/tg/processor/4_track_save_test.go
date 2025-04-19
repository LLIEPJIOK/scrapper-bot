package processor_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTrackSaver_Handle_InvalidObject(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	cache := mocks.NewMockCache(t)

	saver := processor.NewTrackSaver(client, channels, cache)

	state := &processor.State{
		Message:   "any message",
		ChatID:    100,
		MessageID: 0,
		FSMState:  "initial",
		Object:    "invalid object type",
	}

	result := saver.Handle(context.Background(), state)
	assert.Equal(
		t,
		"fail",
		result.NextState.String(),
		"Expected NextState to be 'fail' for invalid object type",
	)
	assert.True(
		t,
		result.IsAutoTransition,
		"Expected IsAutoTransition to be true for invalid object type",
	)
	assert.Equal(t, state, result.Result, "Expected returned state to match the input state")
}

func TestTrackSaver_Handle_ClientError(t *testing.T) {
	t.Parallel()

	link := &domain.Link{
		URL: "http://example.com",
	}
	state := &processor.State{
		Message:   "any message",
		ChatID:    200,
		MessageID: 0,
		FSMState:  "initial",
		Object:    link,
	}
	expErr := errors.New("client add link error")
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("AddLink", context.Background(), link).Return(expErr)

	cache := mocks.NewMockCache(t)

	saver := processor.NewTrackSaver(client, channels, cache)

	result := saver.Handle(context.Background(), state)
	assert.Equal(
		t,
		"fail",
		result.NextState.String(),
		"Expected NextState to be 'fail' when client.AddLink returns error",
	)
	assert.True(
		t,
		result.IsAutoTransition,
		"Expected IsAutoTransition to be true when client.AddLink returns error",
	)
	assert.Equal(
		t,
		"ошибка при добавлении ссылки",
		state.ShowError,
		"Expected ShowError to be set on client error",
	)
	require.Error(t, result.Error, "Expected an error to be returned when client.AddLink fails")
}

func TestTrackSaver_Handle_SuccessNewMessage(t *testing.T) {
	t.Parallel()

	link := &domain.Link{
		URL: "http://example.com",
	}
	state := &processor.State{
		Message:   "any message",
		ChatID:    300,
		MessageID: 0,
		FSMState:  "initial",
		Object:    link,
	}
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("AddLink", context.Background(), link).Return(nil)

	cache := mocks.NewMockCache(t)
	cache.On("InvalidateListLinks", mock.Anything, int64(300)).
		Return(nil).
		Once()

	saver := processor.NewTrackSaver(client, channels, cache)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg edit message")

		expectedText := "Ссылка успешно добавлена!"
		assert.Equal(t, expectedText, msg.Text, "Expected success message text")
	}()

	result := saver.Handle(context.Background(), state)
	assert.Nil(t, state.Object, "Expected state.Object to be nil after handling")
	assert.False(t, result.IsAutoTransition, "Expected IsAutoTransition to be false on success")
	assert.Equal(t, state, result.Result, "Expected returned state to match the input state")

	wg.Wait()
}

func TestTrackSaver_Handle_SuccessEditMessage(t *testing.T) {
	t.Parallel()

	link := &domain.Link{
		URL: "http://example.com",
	}
	state := &processor.State{
		Message:   "any message",
		ChatID:    300,
		MessageID: 123,
		FSMState:  "initial",
		Object:    link,
	}
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("AddLink", context.Background(), link).Return(nil)

	cache := mocks.NewMockCache(t)
	cache.On("InvalidateListLinks", mock.Anything, int64(300)).
		Return(nil).
		Once()

	saver := processor.NewTrackSaver(client, channels, cache)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.EditMessageTextConfig)
		require.True(t, ok, "not tg edit message")

		expectedText := "Ссылка успешно добавлена!"
		assert.Equal(t, expectedText, msg.Text, "Expected success message text")
	}()

	result := saver.Handle(context.Background(), state)
	assert.Nil(t, state.Object, "Expected state.Object to be nil after handling")
	assert.False(t, result.IsAutoTransition, "Expected IsAutoTransition to be false on success")
	assert.Equal(t, state, result.Result, "Expected returned state to match the input state")

	wg.Wait()
}
