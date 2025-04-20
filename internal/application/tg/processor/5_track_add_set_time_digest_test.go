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

func TestTrackAddTimeSetterDigest_Handle_InvalidObject(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	setter := processor.NewTrackAddTimeSetterDigest(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		FSMState:  "track_add_set_time_digest",
		Object:    "invalid object",
	}

	result := setter.Handle(context.Background(), state)

	assert.Equal(t, "fail", result.NextState.String(), "NextState should be fail")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestTrackAddTimeSetterDigest_Handle_SuccessWithKeyboard(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	setter := processor.NewTrackAddTimeSetterDigest(channels)

	link := &domain.Link{
		URL:             "https://example.com",
		ChatID:          123,
		SendImmediately: domain.Null[bool]{},
	}
	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		FSMState:  "track_add_set_time_digest",
		Object:    link,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.EditMessageTextConfig)
		require.True(t, ok, "not tg edit message")

		assert.Equal(
			t,
			"Можете настроить ссылку или сохранить её в текущем состоянии.",
			msg.Text,
			"wrong message text",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match")
		assert.Equal(t, state.MessageID, msg.MessageID, "MessageID should match")
		require.NotNil(t, msg.ReplyMarkup, "ReplyMarkup should be set")
	}()

	result := setter.Handle(context.Background(), state)

	updatedLink, ok := state.Object.(*domain.Link)
	require.True(t, ok, "Expected state.Object to be *domain.Link")
	assert.False(t, updatedLink.SendImmediately.Value, "SendImmediately should be false")

	assert.Equal(t, "callback", result.NextState.String(), "NextState should be callback")
	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the original state")

	wg.Wait()
}

func TestTrackAddTimeSetterDigest_Handle_SuccessWithoutKeyboard(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	setter := processor.NewTrackAddTimeSetterDigest(channels)

	link := &domain.Link{
		URL:             "https://example.com",
		ChatID:          123,
		SendImmediately: domain.Null[bool]{},
		Tags:            []string{"tag1"},
		Filters:         []string{"filter1"},
	}
	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		FSMState:  "track_add_set_time_digest",
		Object:    link,
	}

	result := setter.Handle(context.Background(), state)

	updatedLink, ok := state.Object.(*domain.Link)
	require.True(t, ok, "Expected state.Object to be *domain.Link")
	assert.False(t, updatedLink.SendImmediately.Value, "SendImmediately should be false")

	assert.Equal(t, "track_save", result.NextState.String(), "NextState should be track_save")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}
