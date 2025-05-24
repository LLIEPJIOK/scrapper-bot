package bot_test

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/http/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/http/bot/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	api "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const exampleLink = "https://example.com"

func TestClient_UpdatesPost_InvalidURL(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	botClient := bot.NewClient(clientMock)

	err := botClient.UpdatesPost(context.Background(), &domain.Update{
		URL: "://invalid-url",
	})

	assert.Error(t, err, "expected error")
	assert.Contains(t, err.Error(), "failed to parse link")
}

func TestClient_UpdatesPost_RequestError(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	botClient := bot.NewClient(clientMock)

	testURL := exampleLink
	parsedURL, err := url.Parse(testURL)
	require.NoError(t, err, "failed to parse link")

	chatID := int64(12345)
	tags := []string{"tag1", "tag2"}
	msg := "msg1"

	expectedRequest := &api.LinkUpdate{
		ChatID:          api.NewOptInt64(chatID),
		URL:             api.NewOptURI(*parsedURL),
		Message:         api.NewOptString(msg),
		Tags:            tags,
		SendImmediately: api.NewOptBool(true),
	}
	expectedErr := errors.New("network error")

	clientMock.On("UpdatesPost", mock.Anything, expectedRequest).Return(nil, expectedErr).Once()

	err = botClient.UpdatesPost(context.Background(), &domain.Update{
		ChatID:          chatID,
		URL:             testURL,
		Message:         msg,
		Tags:            tags,
		SendImmediately: domain.NewNull(true),
	})

	require.Error(t, err, "expected error")
	assert.ErrorAs(t, err, &client.ErrServiceUnavailable{}, "should be ErrServiceUnavailable")
	assert.Contains(t, err.Error(), "failed to send updates")
}

func TestClient_UpdatesPost_Success(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	botClient := bot.NewClient(clientMock)

	testURL := exampleLink
	parsedURL, err := url.Parse(testURL)
	require.NoError(t, err, "failed to parse link")

	chatID := int64(12345)
	tags := []string{"tag1", "tag2"}
	msg := "msg2"

	expectedRequest := &api.LinkUpdate{
		ChatID:          api.NewOptInt64(chatID),
		URL:             api.NewOptURI(*parsedURL),
		Message:         api.NewOptString(msg),
		Tags:            tags,
		SendImmediately: api.NewOptBool(false),
	}
	clientMock.On("UpdatesPost", mock.Anything, expectedRequest).
		Return(&api.UpdatesPostOK{}, nil).
		Once()

	err = botClient.UpdatesPost(context.Background(), &domain.Update{
		ChatID:          chatID,
		URL:             testURL,
		Message:         msg,
		Tags:            tags,
		SendImmediately: domain.NewNull(false),
	})

	assert.NoError(t, err, "expected no error")
}

func TestClient_UpdatesPost_ApiErrorResponse(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	botClient := bot.NewClient(clientMock)

	testURL := exampleLink
	parsedURL, err := url.Parse(testURL)
	require.NoError(t, err, "failed to parse link")

	chatID := int64(12345)
	tags := []string{"tag1", "tag2"}
	msg := "msg3"

	expectedRequest := &api.LinkUpdate{
		ChatID:          api.NewOptInt64(chatID),
		URL:             api.NewOptURI(*parsedURL),
		Message:         api.NewOptString(msg),
		Tags:            tags,
		SendImmediately: api.NewOptBool(false),
	}

	apiErr := &api.ApiErrorResponse{
		Description: api.NewOptString("invalid link"),
	}

	clientMock.On("UpdatesPost", mock.Anything, expectedRequest).Return(apiErr, nil).Once()

	err = botClient.UpdatesPost(context.Background(), &domain.Update{
		ChatID:  chatID,
		URL:     testURL,
		Message: msg,
		Tags:    tags,
	})

	require.Error(t, err, "expected error")
	assert.ErrorAs(t, err, &client.ErrServiceUnavailable{}, "should be ErrServiceUnavailable")
	assert.Contains(t, err.Error(), "failed to add link: invalid link")
}
