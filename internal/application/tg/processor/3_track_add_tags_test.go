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

func TestTrackTagAdder_Handle_InvalidObject(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	adder := processor.NewTrackTagAdder(channels)

	state := &processor.State{
		Message:  "filter1 filter2",
		ChatID:   100,
		FSMState: "initial",
		Object:   "invalid object",
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
	assert.Equal(t, state, result.Result, "Expected returned state to equal input state")
}

func TestTrackTagAdder_Handle_CallbackBranch(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	adder := processor.NewTrackTagAdder(channels)

	link := &domain.Link{
		Filters:         nil,
		Tags:            nil,
		SendImmediately: domain.NewNull(true),
	}
	state := &processor.State{
		Message:  "tag1 tag2",
		ChatID:   200,
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

		expectedText := "Можете настроить ссылку или сохранить её в текущем состоянии."
		assert.Equal(t, expectedText, msg.Text, "Expected message text to match")
		assert.NotNil(t, msg.ReplyMarkup, "Expected message ReplyMarkup to be set")
	}()

	result := adder.Handle(context.Background(), state)

	updatedLink, ok := state.Object.(*domain.Link)
	require.True(t, ok, "Expected state.Object to be *domain.Link")

	expectedTags := strings.Fields(state.Message)
	assert.Equal(
		t,
		expectedTags,
		updatedLink.Tags,
		"Expected link.Tags to be updated based on state.Message",
	)

	assert.Equal(t, "callback", result.NextState.String(), "Expected NextState to be 'callback'")
	assert.False(t, result.IsAutoTransition, "Expected IsAutoTransition to be false")
	assert.Equal(t, state, result.Result, "Expected returned state to equal input state")

	wg.Wait()
}

func TestTrackTagAdder_Handle_TrackSaveBranch(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	adder := processor.NewTrackTagAdder(channels)

	link := &domain.Link{
		Filters:         []string{"filter1", "filter2"},
		Tags:            nil,
		SendImmediately: domain.NewNull(true),
	}
	state := &processor.State{
		Message:  "tag1 tag2",
		ChatID:   300,
		FSMState: "initial",
		Object:   link,
	}

	result := adder.Handle(context.Background(), state)

	updatedLink, ok := state.Object.(*domain.Link)
	require.True(t, ok, "Expected state.Object to be *domain.Link")

	expectedTags := strings.Fields(state.Message)
	assert.Equal(
		t,
		expectedTags,
		updatedLink.Tags,
		"Expected link.Tags to be updated based on state.Message",
	)

	assert.Equal(t, "track_save", result.NextState.String(), "Expected NextState to be 'trackSave'")
	assert.True(t, result.IsAutoTransition, "Expected IsAutoTransition to be true")
	assert.Equal(t, state, result.Result, "Expected returned state to equal input state")
}
