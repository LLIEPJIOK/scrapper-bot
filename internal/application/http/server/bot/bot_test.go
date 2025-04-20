package bot_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/server/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/server/bot/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	api "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleURL = "https://example.com"

func TestServer_UpdatesPost_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repoMock := mocks.NewMockRepository(t)
	channels := domain.NewChannels()
	server := bot.NewServer(repoMock, channels)

	parsedURL, err := url.Parse(exampleURL)
	require.NoError(t, err, "url parse error")

	tags := []string{"tag1", "tag2"}
	msg := "msg1"
	chatID := int64(12345)

	req := &api.LinkUpdate{
		URL:             api.NewOptURI(*parsedURL),
		ChatID:          api.NewOptInt64(chatID),
		Message:         api.NewOptString(msg),
		Tags:            tags,
		SendImmediately: api.NewOptBool(false),
	}

	repoMock.On("AddUpdate", ctx, &domain.Update{
		ChatID:  chatID,
		URL:     exampleURL,
		Message: msg,
		Tags:    tags,
	}).Return(nil).Once()

	res, err := server.UpdatesPost(ctx, req)

	assert.NoError(t, err, "server error")
	assert.IsType(t, &api.UpdatesPostOK{}, res, "response type error")
}

func TestServer_UpdatesPost_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repoMock := mocks.NewMockRepository(t)
	channels := domain.NewChannels()
	server := bot.NewServer(repoMock, channels)

	parsedURL, err := url.Parse(exampleURL)
	require.NoError(t, err, "url parse error")

	tags := []string{"tag1", "tag2"}
	msg := "msg"
	chatID := int64(12345)

	req := &api.LinkUpdate{
		URL:             api.NewOptURI(*parsedURL),
		ChatID:          api.NewOptInt64(chatID),
		Message:         api.NewOptString(msg),
		Tags:            tags,
		SendImmediately: api.NewOptBool(false),
	}

	repoMock.On("AddUpdate", ctx, &domain.Update{
		ChatID:  chatID,
		URL:     exampleURL,
		Message: msg,
		Tags:    tags,
	}).Return(errors.New("database error")).Once()

	res, err := server.UpdatesPost(ctx, req)

	assert.Error(t, err, "server error")
	assert.IsType(t, &api.ApiErrorResponse{}, res, "response type error")
	assert.Equal(
		t,
		http.StatusText(http.StatusInternalServerError),
		res.(*api.ApiErrorResponse).Code.Value,
		"response code error",
	)
}

func TestServer_ImmediatelyUpdatesPost_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repoMock := mocks.NewMockRepository(t)
	channels := domain.NewChannels()
	server := bot.NewServer(repoMock, channels)

	parsedURL, err := url.Parse(exampleURL)
	require.NoError(t, err, "url parse error")

	tags := []string{"tag1", "tag2"}
	msg := "msg"
	chatID := int64(12345)

	req := &api.LinkUpdate{
		URL:             api.NewOptURI(*parsedURL),
		ChatID:          api.NewOptInt64(chatID),
		Message:         api.NewOptString(msg),
		Tags:            tags,
		SendImmediately: api.NewOptBool(true),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		resp := <-channels.TelegramResp()

		tgMessage, ok := resp.(tgbotapi.MessageConfig)
		require.True(t, ok, "invalid message type")
		assert.Equal(t, msg, tgMessage.Text, "message error")
	}()

	res, err := server.UpdatesPost(ctx, req)

	assert.NoError(t, err, "server error")
	assert.IsType(t, &api.UpdatesPostOK{}, res, "response type error")

	wg.Wait()
}
