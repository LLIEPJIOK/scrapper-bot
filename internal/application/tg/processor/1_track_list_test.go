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

func TestHandle_TrackLister_NoLinks(t *testing.T) {
	t.Parallel()

	client := mocks.NewMockClient(t)
	client.On("GetLinks", mock.Anything, int64(123)).Return([]*domain.Link{}, nil).Once()

	channels := domain.NewChannels()

	trackLister := processor.NewTrackLister(client, channels)
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

		expectedText := "У вас нет ни одной ссылки. Для добавления ссылки воспользуйтесь командой /track"
		assert.Equal(
			t,
			expectedText,
			msg.Text,
			"Text should show no links",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should be the same")
	}()

	result := trackLister.Handle(context.Background(), state)

	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(t, state, result.Result, "Result should be the same as the state")
	assert.Nil(t, result.Error, "Error should be nil")

	wg.Wait()
}

func TestHandle_TrackLister_LinksWithoutTagsOrFilters(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	links := []*domain.Link{
		{URL: "https://example.com"},
		{URL: "https://test.com"},
	}
	client := mocks.NewMockClient(t)
	client.On("GetLinks", mock.Anything, int64(123)).Return(links, nil).Once()

	trackLister := processor.NewTrackLister(client, channels)
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

		expectedText := `Ваши ссылки:
1) https://example.com

2) https://test.com

`
		assert.Equal(
			t,
			expectedText,
			msg.Text,
			"Text should be without tags and filters",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should be the same")
		assert.Equal(t, tgbotapi.ModeMarkdown, msg.ParseMode, "ParseMode should be Markdown")
	}()

	result := trackLister.Handle(context.Background(), state)

	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should be the same as the state")
	assert.Nil(t, result.Error, "Error should be nil")

	wg.Wait()
}

func TestHandle_TrackLister_LinksWithTagsAndFilters(t *testing.T) {
	t.Parallel()

	channels := domain.NewChannels()
	links := []*domain.Link{
		{URL: "https://example.com", Tags: []string{"tag1", "tag2"}, Filters: []string{"filter1"}},
		{URL: "https://test.com", Tags: []string{"tag3"}, Filters: []string{"filter2", "filter3"}},
	}
	client := mocks.NewMockClient(t)
	client.On("GetLinks", mock.Anything, int64(123)).Return(links, nil).Once()

	trackLister := processor.NewTrackLister(client, channels)
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

		expectedText := `Ваши ссылки:
1) https://example.com
*Тэги:* tag1; tag2
*Фильтры:* filter1

2) https://test.com
*Тэги:* tag3
*Фильтры:* filter2; filter3

`
		assert.Equal(
			t,
			expectedText,
			msg.Text,
			"Text should be with tags and filters",
		)
		assert.Equal(t, state.ChatID, msg.ChatID, "ChatID should be the same")
		assert.Equal(t, tgbotapi.ModeMarkdown, msg.ParseMode, "ParseMode should be Markdown")
	}()

	result := trackLister.Handle(context.Background(), state)

	assert.False(t, result.IsAutoTransition, "IsAutoTransition should be false")
	assert.Equal(t, state, result.Result, "Result should be the same as the state")
	assert.Nil(t, result.Error, "Error should be nil")

	wg.Wait()
}

func TestHandle_TrackLister_GetLinksError(t *testing.T) {
	t.Parallel()

	getLinksErr := errors.New("failed to get links")
	client := mocks.NewMockClient(t)
	client.On("GetLinks", mock.Anything, int64(123)).Return(nil, getLinksErr).Once()

	channels := domain.NewChannels()

	trackLister := processor.NewTrackLister(client, channels)
	state := &processor.State{
		ChatID: 123,
	}

	result := trackLister.Handle(context.Background(), state)

	assert.Equal(t, "fail", result.NextState.String(), "NextState should be fail")
	assert.True(t, result.IsAutoTransition, "IsAutoTransition should be true")
	assert.Equal(
		t,
		"не удалось получить ссылки",
		result.Result.ShowError,
		"ShowError should exists",
	)
	assert.Equal(t, state, result.Result, "Result should be the same as the state")
	assert.NotNil(t, result.Error, "Ошибка не должна быть nil")
	assert.Contains(
		t,
		result.Error.Error(),
		"h.client.GetLinks(ctx, 123): failed to get links",
		"Error should contains get links error",
	)
}
