package processor_test

import (
	"context"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"
)

func TestByTagInputLister_Handle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	channels := domain.NewChannels()
	lister := processor.NewByTagInputLister(channels)

	state := &processor.State{
		ChatID:    12345,
		MessageID: 67890,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		resp := <-channels.TelegramResp()

		msg, ok := resp.(tgbotapi.EditMessageTextConfig)
		require.True(t, ok, "not tg edit message")
		require.Equal(t, state.ChatID, msg.ChatID, "ChatID should match the state's ChatID")
		require.Equal(
			t,
			state.MessageID,
			msg.MessageID,
			"MessageID should match the state's MessageID",
		)
		require.Equal(t, "Введите тег", msg.Text, "Message text should be 'Введите тег'")
	}()

	result := lister.Handle(ctx, state)

	require.Equal(t, "list_by_tag", result.NextState.String(), "NextState should be ListByTag")
	require.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	require.Equal(t, state, result.Result, "Result should equal the original state")
}
