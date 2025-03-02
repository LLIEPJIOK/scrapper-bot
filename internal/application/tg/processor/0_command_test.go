package processor_test

import (
	"context"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/stretchr/testify/assert"
)

func TestHandleStartCommand(t *testing.T) {
	t.Parallel()

	commander := processor.NewCommander()
	state := &processor.State{Message: "/start"}
	result := commander.Handle(context.Background(), state)

	assert.Equal(t, "start", result.NextState.String(), "NextState should be start")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandleHelpCommand(t *testing.T) {
	t.Parallel()

	commander := processor.NewCommander()
	state := &processor.State{Message: "/help"}
	result := commander.Handle(context.Background(), state)

	assert.Equal(t, "help", result.NextState.String(), "NextState should be help")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandleTrackCommand(t *testing.T) {
	t.Parallel()

	commander := processor.NewCommander()
	state := &processor.State{Message: "/track"}
	result := commander.Handle(context.Background(), state)

	assert.Equal(t, "track", result.NextState.String(), "NextState should be track")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandle_UntrackCommand(t *testing.T) {
	t.Parallel()

	commander := processor.NewCommander()
	state := &processor.State{Message: "/untrack"}
	result := commander.Handle(context.Background(), state)

	assert.Equal(t, "untrack", result.NextState.String(), "NextState should be untrack")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandle_ListCommand(t *testing.T) {
	t.Parallel()

	commander := processor.NewCommander()
	state := &processor.State{Message: "/list"}
	result := commander.Handle(context.Background(), state)

	assert.Equal(t, "list", result.NextState.String(), "NextState should be trackList")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should contain the original state")
}

func TestHandleUnknownCommand(t *testing.T) {
	t.Parallel()

	commander := processor.NewCommander()
	state := &processor.State{Message: "/unknown"}
	result := commander.Handle(context.Background(), state)

	assert.Equal(t, "fail", result.NextState.String(), "NextState should be fail")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(
		t,
		"неопознанная команда",
		result.Result.ShowError,
		"ShowError should be set to 'неопознанная команда'",
	)
	assert.Equal(t, state, result.Result, "Result should contain the modified state")
}
