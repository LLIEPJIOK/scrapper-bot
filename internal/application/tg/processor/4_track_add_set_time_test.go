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

func TestTrackAddTimeSetter_Handle_Success(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	setter := processor.NewTrackAddTimeSetter(channels)

	state := &processor.State{
		ChatID:    123,
		MessageID: 456,
		FSMState:  "track_add_set_time",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.EditMessageTextConfig)
		require.True(t, ok, "not tg edit message")

		assert.Equal(t, processor.TrackAddTimeSetterText, msg.Text, "wrong message text")
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match")
		assert.Equal(t, state.MessageID, msg.MessageID, "MessageID should match")
		assert.Equal(t, tgbotapi.ModeMarkdown, msg.ParseMode, "ParseMode should be Markdown")

		require.NotNil(t, msg.ReplyMarkup, "ReplyMarkup should be set")
		require.Len(t, msg.ReplyMarkup.InlineKeyboard, 2, "should have 2 rows of buttons")

		require.Len(t, msg.ReplyMarkup.InlineKeyboard[0], 1, "first row should have 1 button")
		assert.Equal(
			t,
			"По расписанию",
			msg.ReplyMarkup.InlineKeyboard[0][0].Text,
			"first button text should be 'По расписанию'",
		)

		require.Len(t, msg.ReplyMarkup.InlineKeyboard[1], 1, "second row should have 1 button")
		assert.Equal(
			t,
			"Сразу",
			msg.ReplyMarkup.InlineKeyboard[1][0].Text,
			"second button text should be 'Сразу'",
		)
	}()

	result := setter.Handle(context.Background(), state)

	assert.Equal(t, "callback", result.NextState.String(), "NextState should be callback")
	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the original state")

	wg.Wait()
}
