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

func TestHandleTrackAddTags(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	callbacker := processor.NewCallbacker(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		Message:   "track_add_tags",
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
			"Введите теги через пробел.",
			msg.Text,
			"wrong message",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match")
		assert.Equal(t, state.MessageID, msg.MessageID, "MessageID should match")
	}()

	result := callbacker.Handle(context.Background(), state)

	assert.Equal(t, "track_add_tags", result.NextState.String(), "NextState should be trackAddTags")
	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the original state")

	wg.Wait()
}

func TestHandleTrackAddFilters(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	callbacker := processor.NewCallbacker(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		Message:   "track_add_filters",
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
			"Введите фильтры через пробел.\nПример: user:dummy type:comment",
			msg.Text,
			"wrong message",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match")
		assert.Equal(t, state.MessageID, msg.MessageID, "MessageID should match")
	}()

	result := callbacker.Handle(context.Background(), state)

	assert.Equal(
		t,
		"track_add_filters",
		result.NextState.String(),
		"NextState should be trackAddFilters",
	)
	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the original state")

	wg.Wait()
}

func TestHandleTrackSave(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	callbacker := processor.NewCallbacker(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		Message:   "track_save",
	}

	result := callbacker.Handle(context.Background(), state)

	assert.Equal(t, "track_save", result.NextState.String(), "NextState should be trackSave")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandleListAll(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	callbacker := processor.NewCallbacker(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		Message:   "list_all",
	}

	result := callbacker.Handle(context.Background(), state)

	assert.Equal(t, "list_all", result.NextState.String(), "NextState should be list_all")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandleListByTagInput(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	callbacker := processor.NewCallbacker(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		Message:   "list_by_tag_input",
	}

	result := callbacker.Handle(context.Background(), state)

	assert.Equal(
		t,
		"list_by_tag_input",
		result.NextState.String(),
		"NextState should be list_by_tag_input",
	)
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandleUnknownCallback(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	callbacker := processor.NewCallbacker(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		Message:   "unknown",
	}

	result := callbacker.Handle(context.Background(), state)

	assert.Equal(t, "fail", result.NextState.String(), "NextState should be fail")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, "неопознанная команда", result.Result.ShowError, "ShowError should be set")
	assert.Equal(t, state, result.Result, "Result should contain the modified state")
}
