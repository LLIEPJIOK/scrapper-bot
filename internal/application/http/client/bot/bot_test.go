package bot_test

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/client/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/client/bot/mocks"
	api "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const exampleLink = "https://example.com"

func TestClient_UpdatesPost_InvalidURL(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := bot.NewClient(clientMock)

	err := client.UpdatesPost(context.Background(), "://invalid-url", []int64{12345})

	assert.Error(t, err, "expected error")
	assert.Contains(t, err.Error(), "failed to parse link")
}

func TestClient_UpdatesPost_RequestError(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := bot.NewClient(clientMock)

	testURL := exampleLink
	parsedURL, err := url.Parse(testURL)
	require.NoError(t, err, "failed to parse link")

	expectedRequest := &api.LinkUpdate{
		URL:       api.NewOptURI(*parsedURL),
		TgChatIds: []int64{12345},
	}
	expectedErr := errors.New("network error")

	clientMock.On("UpdatesPost", mock.Anything, expectedRequest).Return(nil, expectedErr).Once()

	err = client.UpdatesPost(context.Background(), testURL, []int64{12345})

	assert.Error(t, err, "expected error")
	assert.Contains(t, err.Error(), "failed to send updates")
}

func TestClient_UpdatesPost_Success(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := bot.NewClient(clientMock)

	testURL := exampleLink
	parsedURL, err := url.Parse(testURL)
	require.NoError(t, err, "failed to parse link")

	expectedRequest := &api.LinkUpdate{
		URL:       api.NewOptURI(*parsedURL),
		TgChatIds: []int64{12345},
	}

	clientMock.On("UpdatesPost", mock.Anything, expectedRequest).
		Return(&api.UpdatesPostOK{}, nil).
		Once()

	err = client.UpdatesPost(context.Background(), testURL, []int64{12345})

	assert.NoError(t, err, "expected no error")
}

func TestClient_UpdatesPost_ApiErrorResponse(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := bot.NewClient(clientMock)

	testURL := exampleLink
	parsedURL, err := url.Parse(testURL)
	require.NoError(t, err, "failed to parse link")

	expectedRequest := &api.LinkUpdate{
		URL:       api.NewOptURI(*parsedURL),
		TgChatIds: []int64{12345},
	}

	apiErr := &api.ApiErrorResponse{
		Description: api.NewOptString("invalid link"),
	}

	clientMock.On("UpdatesPost", mock.Anything, expectedRequest).Return(apiErr, nil).Once()

	err = client.UpdatesPost(context.Background(), testURL, []int64{12345})

	assert.Error(t, err, "expected error")
	assert.Contains(t, err.Error(), "failed to add link: invalid link")
}
