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

func TestHandle_Untracker(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	untracker := processor.NewUntracker(channels)

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

		expectedText := "Введите ссылку на ресурс, который хотите удалить."
		assert.Equal(
			t,
			expectedText,
			msg.Text,
			"Message text should match the expected untracking prompt",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match the state's ChatID")
	}()

	result := untracker.Handle(context.Background(), state)

	assert.Equal(
		t,
		"untrack_delete_link",
		result.NextState.String(),
		"NextState should be untrackDeleteLink",
	)
	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the updated state")

	assert.NotNil(t, state.Object, "Object should not be nil")
	link, ok := state.Object.(*domain.Link)
	assert.True(t, ok, "Object should be of type *domain.Link")
	assert.Equal(t, state.ChatID, link.ChatID, "Link.ChatID should match state.ChatID")

	wg.Wait()
}
