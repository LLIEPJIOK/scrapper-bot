package bot_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/server/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/server/bot/mocks"
	api "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleURL = "https://example.com"

func TestServer_UpdatesPost_Success(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	server := bot.NewServer(repoMock)

	parsedURL, err := url.Parse(exampleURL)
	require.NoError(t, err, "url parse error")

	req := &api.LinkUpdate{
		TgChatIds: []int64{123, 456},
		URL:       api.NewOptURI(*parsedURL),
	}

	repoMock.On("AddLink", int64(123), exampleURL).Return(nil).Once()
	repoMock.On("AddLink", int64(456), exampleURL).Return(nil).Once()

	res, err := server.UpdatesPost(context.Background(), req)

	assert.NoError(t, err, "server error")
	assert.IsType(t, &api.UpdatesPostOK{}, res, "response type error")
}

func TestServer_UpdatesPost_Error(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	server := bot.NewServer(repoMock)

	parsedURL, err := url.Parse(exampleURL)
	require.NoError(t, err, "url parse error")

	req := &api.LinkUpdate{
		TgChatIds: []int64{123},
		URL:       api.NewOptURI(*parsedURL),
	}

	repoMock.On("AddLink", int64(123), exampleURL).
		Return(errors.New("database error")).
		Once()

	res, err := server.UpdatesPost(context.Background(), req)

	assert.Error(t, err, "server error")
	assert.IsType(t, &api.ApiErrorResponse{}, res, "response type error")
	assert.Equal(
		t,
		http.StatusText(http.StatusInternalServerError),
		res.(*api.ApiErrorResponse).Code.Value,
		"response code error",
	)
}
