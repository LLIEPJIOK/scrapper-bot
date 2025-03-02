package processor_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrackFilterAdder_Handle_InvalidObject(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	adder := processor.NewTrackFilterAdder(channels)

	state := &processor.State{
		Message:  "filter1 filter2",
		ChatID:   123,
		FSMState: "initial",
		Object:   "invalid",
	}

	result := adder.Handle(context.Background(), state)

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
	assert.Equal(
		t,
		state,
		result.Result,
		"Expected the returned state to be the same as the input state",
	)
}

func TestTrackFilterAdder_Handle(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	adder := processor.NewTrackFilterAdder(channels)

	link := &domain.Link{
		Filters: nil,
	}
	state := &processor.State{
		Message:  "filter1 filter2 filter3",
		ChatID:   456,
		FSMState: "initial",
		Object:   link,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg edit message")

		expectedText := "Можете добавить опциональные поля или сохранить ссылку в текущем состоянии."
		assert.Equal(t, expectedText, msg.Text, "Expected message text to match")
		assert.NotNil(t, msg.ReplyMarkup, "Expected ReplyMarkup to be set in the message")
	}()

	result := adder.Handle(context.Background(), state)

	updatedLink, ok := state.Object.(*domain.Link)
	require.True(t, ok, "Expected state.Object to be of type *domain.Link")

	expectedFilters := strings.Fields(state.Message)
	assert.Equal(
		t,
		expectedFilters,
		updatedLink.Filters,
		"Expected link.Filters to be updated based on state.Message",
	)

	assert.Equal(
		t,
		"callback",
		result.NextState.String(),
		"Expected NextState to be 'callback' on success",
	)
	assert.False(t, result.IsAutoTransition, "Expected IsAutoTransition to be false on success")
	assert.Equal(
		t,
		state,
		result.Result,
		"Expected the returned state to be the same as the input state",
	)

	wg.Wait()
}
