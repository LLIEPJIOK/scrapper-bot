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

func TestHandleHelper(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	helper := processor.NewHelper(channels)

	state := &processor.State{
		ChatID: 123,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg edit message")
		assert.Equal(t, `📌 Доступные команды:  
- /track – подписаться на обновления
- /track – отписаться от обновлений
- /list – показать все подписки
- /help – справка по командам
`, msg.Text, "Message text should match helperAnswer")
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match the state's ChatID")
	}()

	result := helper.Handle(context.Background(), state)

	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the original state")

	wg.Wait()
}
