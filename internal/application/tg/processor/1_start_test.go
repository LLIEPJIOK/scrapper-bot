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

func TestHandleStater(t *testing.T) {
	t.Parallel()

	client := mocks.NewMockClient(t)
	client.On("RegisterChat", mock.Anything, int64(123)).Return(nil)

	channels := domain.NewChannels()

	stater := processor.NewStater(client, channels)
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
		assert.Equal(
			t,
			`*Привет! Я LinkTracker – твой помощник для отслеживания обновлений на сайтах.*  

🔹 Подписывайся на ссылки и получай уведомления об изменениях.  
🔹 Управляй подписками прямо в Telegram.  
🔹 Получай обновления сразу или в удобное время.  

📌 Доступные команды:  
- /track – подписаться на обновления
- /untrack – отписаться от обновлений
- /list – показать все подписки
- /help – справка по командам

Начни с /track и будь в курсе важных событий! 🚀
`,
			msg.Text,
			"Message text should match staterAnswer",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match the state's ChatID")
		assert.Equal(t, tgbotapi.ModeMarkdown, msg.ParseMode, "ParseMode should be Markdown")
	}()

	result := stater.Handle(context.Background(), state)

	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
	assert.Nil(t, result.Error, "Error should be nil")

	wg.Wait()
}

func TestHandleStaterWithRegistrationError(t *testing.T) {
	t.Parallel()

	registrationErr := errors.New("registration failed")
	channels := domain.NewChannels()

	client := mocks.NewMockClient(t)
	client.On("RegisterChat", mock.Anything, int64(123)).Return(registrationErr)

	stater := processor.NewStater(client, channels)
	state := &processor.State{
		ChatID: 123,
	}

	result := stater.Handle(context.Background(), state)

	assert.Equal(t, "fail", result.NextState.String(), "NextState should be fail")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, "ошибка регистрации чата", result.Result.ShowError, "ShowError should be set")
	assert.Equal(t, state, result.Result, "Result should contain the modified state")
	assert.NotNil(t, result.Error, "Error should not be nil")
	assert.Contains(
		t,
		result.Error.Error(),
		"h.client.RegisterChat(ctx, 123): registration failed",
		"Error message should contain registration error",
	)
}
