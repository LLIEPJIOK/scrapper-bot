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

func TestListerHandle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	channels := domain.NewChannels()
	lister := processor.NewLister(channels)

	state := &processor.State{
		ChatID: 12345,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		resp := <-channels.TelegramResp()
		msg, ok := resp.(tgbotapi.MessageConfig)
		require.True(t, ok, "invalid message type")

		assert.Equal(t, state.ChatID, msg.ChatID, "invalid ChatID")
		assert.Equal(t, "Какие ссылки вы хотите получить?", msg.Text, "invalid message text")

		markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		require.True(t, ok, "ReplyMarkup should be InlineKeyboardMarkup")
		require.Len(t, markup.InlineKeyboard, 2, "invalid markup len")

		btn1 := markup.InlineKeyboard[0][0]
		assert.Equal(t, "Все", btn1.Text, "invalid text for first button")
		assert.Equal(t, "list_all", *btn1.CallbackData, "invalid data for first button")

		btn2 := markup.InlineKeyboard[1][0]
		assert.Equal(t, "С определённым тегом", btn2.Text, "invalid text for second button")
		assert.Equal(
			t,
			"list_by_tag_input",
			*btn2.CallbackData,
			"invalid data for second button",
		)
	}()

	result := lister.Handle(ctx, state)
	assert.Equal(t, "callback", result.NextState.String(), "invalid NextState")
	assert.Equal(t, state, result.Result, "invalid Result")
	assert.False(t, result.IsAutoTransition, "invalid IsAutoTransition")

	wg.Wait()
}
