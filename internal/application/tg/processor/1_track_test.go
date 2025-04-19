package processor_test

import (
	"context"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

func TestHandle_Tracker(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	tracker := processor.NewTracker(channels)

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
		assert.Equal(t, `Введите ссылку на ресурс, который хотите отслеживать.
Доступные сайты:
	- GitHub
	- StackOverflow
`, msg.Text, "Message text should match trackerAnswer")
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should match the state's ChatID")
	}()

	result := tracker.Handle(context.Background(), state)

	assert.Equal(t, "track_add_link", result.NextState.String(), "NextState should be trackAddLink")
	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should contain the updated state")

	assert.NotNil(t, state.Object, "Object should not be nil")
	link, ok := state.Object.(*domain.Link)
	assert.True(t, ok, "Object should be of type *domain.Link")
	assert.Equal(t, state.ChatID, link.ChatID, "Link.ChatID should match state.ChatID")

	wg.Wait()
}
