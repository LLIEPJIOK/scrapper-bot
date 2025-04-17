package scrapper_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/client/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/http/client/scrapper/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	api "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const exampleLink = "https://example.com"

func TestClient_RegisterChat_Success(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)

	clientMock.On("TgChatIDPost", mock.Anything, api.TgChatIDPostParams{ID: chatID}).
		Return(&api.TgChatIDPostOK{}, nil).Once()

	err := client.RegisterChat(context.Background(), chatID)

	assert.NoError(t, err, "RegisterChat should not return error")
}

func TestClient_RegisterChat_Error(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)
	expectedErr := errors.New("network error")

	clientMock.On("TgChatIDPost", mock.Anything, api.TgChatIDPostParams{ID: chatID}).
		Return(nil, expectedErr).Once()

	err := client.RegisterChat(context.Background(), chatID)

	assert.Error(t, err, "RegisterChat should return error")
	assert.Contains(t, err.Error(), "failed to register chat", "RegisterChat should return error")
}

func TestClient_RegisterChat_APIError(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)

	clientMock.On("TgChatIDPost", mock.Anything, api.TgChatIDPostParams{ID: chatID}).
		Return(&api.ApiErrorResponse{
			Description: api.NewOptString("error description"),
		}, nil).Once()

	err := client.RegisterChat(context.Background(), chatID)

	errResp := scrapper.ErrResponse{}
	require.True(t, errors.As(err, &errResp), "RegisterChat should return error response")
	assert.Equal(
		t,
		"failed to register chat: error description",
		errResp.Message,
		"RegisterChat should return error response",
	)
}

func TestClient_AddLink_Success(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	link := &domain.Link{
		URL:     exampleLink,
		Tags:    []string{"news"},
		Filters: []string{"filter1"},
		ChatID:  12345,
	}

	parsedURL, _ := url.Parse(link.URL)
	expectedRequest := &api.AddLinkRequest{
		Link:    api.NewOptURI(*parsedURL),
		Tags:    link.Tags,
		Filters: link.Filters,
	}

	clientMock.On("LinksPost", mock.Anything, expectedRequest, api.LinksPostParams{TgChatID: link.ChatID}).
		Return(&api.LinkResponse{}, nil).
		Once()

	err := client.AddLink(context.Background(), link)

	assert.NoError(t, err, "AddLink should not return error")
}

func TestClient_AddLink_Error(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	link := &domain.Link{
		URL:     exampleLink,
		Tags:    []string{"news"},
		Filters: []string{"filter1"},
		ChatID:  12345,
	}

	parsedURL, _ := url.Parse(link.URL)
	expectedRequest := &api.AddLinkRequest{
		Link:    api.NewOptURI(*parsedURL),
		Tags:    link.Tags,
		Filters: link.Filters,
	}

	clientMock.On("LinksPost", mock.Anything, expectedRequest, api.LinksPostParams{TgChatID: link.ChatID}).
		Return(&api.LinkResponse{}, errors.New("error")).
		Once()

	err := client.AddLink(context.Background(), link)

	assert.EqualError(t, err, "failed to add link: error", "AddLink should return error")
}

func TestClient_AddLink_APIError(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	link := &domain.Link{
		URL:     exampleLink,
		Tags:    []string{"news"},
		Filters: []string{"filter1"},
		ChatID:  12345,
	}
	parsedURL, _ := url.Parse(link.URL)
	expectedRequest := &api.AddLinkRequest{
		Link:    api.NewOptURI(*parsedURL),
		Tags:    link.Tags,
		Filters: link.Filters,
	}

	clientMock.On("LinksPost", mock.Anything, expectedRequest, api.LinksPostParams{TgChatID: link.ChatID}).
		Return(&api.ApiErrorResponse{
			Description: api.NewOptString("error description"),
		}, nil).
		Once()

	err := client.AddLink(context.Background(), link)

	errResp := scrapper.ErrResponse{}
	require.True(t, errors.As(err, &errResp), "RegisterChat should return error response")
	assert.Equal(
		t,
		"failed to add link: error description",
		errResp.Message,
		"RegisterChat should return error response",
	)
}

func TestClient_DeleteLink_Success(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)
	linkURL := exampleLink
	parsedURL, _ := url.Parse(linkURL)
	expectedRequest := &api.RemoveLinkRequest{Link: api.NewOptURI(*parsedURL)}

	clientMock.On("LinksDelete", mock.Anything, expectedRequest, api.LinksDeleteParams{TgChatID: chatID}).
		Return(&api.LinkResponse{}, nil).
		Once()

	err := client.DeleteLink(context.Background(), chatID, linkURL)

	assert.NoError(t, err, "DeleteLink should not return error")
}

func TestClient_DeleteLink_Error(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)
	linkURL := exampleLink
	parsedURL, _ := url.Parse(linkURL)
	expectedRequest := &api.RemoveLinkRequest{Link: api.NewOptURI(*parsedURL)}

	clientMock.On("LinksDelete", mock.Anything, expectedRequest, api.LinksDeleteParams{TgChatID: chatID}).
		Return(&api.LinkResponse{}, errors.New("error")).
		Once()

	err := client.DeleteLink(context.Background(), chatID, linkURL)

	assert.EqualError(t, err, "failed to delete link: error", "DeleteLink should return error")
}

func TestClient_DeleteLink_APIError(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)
	linkURL := exampleLink
	parsedURL, _ := url.Parse(linkURL)
	expectedRequest := &api.RemoveLinkRequest{Link: api.NewOptURI(*parsedURL)}

	clientMock.On("LinksDelete", mock.Anything, expectedRequest, api.LinksDeleteParams{TgChatID: chatID}).
		Return(&api.LinksDeleteBadRequest{
			Description: api.NewOptString("error description"),
		}, nil).
		Once()

	err := client.DeleteLink(context.Background(), chatID, linkURL)

	errResp := scrapper.ErrResponse{}
	require.True(t, errors.As(err, &errResp), "RegisterChat should return error response")
	assert.Equal(
		t,
		"bad request: error description",
		errResp.Message,
		"RegisterChat should return error response",
	)
}

func TestClient_DeleteLink_NotFound(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)
	linkURL := exampleLink
	parsedURL, _ := url.Parse(linkURL)
	expectedRequest := &api.RemoveLinkRequest{Link: api.NewOptURI(*parsedURL)}

	clientMock.On("LinksDelete", mock.Anything, expectedRequest, api.LinksDeleteParams{TgChatID: chatID}).
		Return(&api.LinksDeleteNotFound{Description: api.NewOptString("not found")}, nil).
		Once()

	err := client.DeleteLink(context.Background(), chatID, linkURL)

	assert.Error(t, err, "DeleteLink should return error")
	assert.Contains(
		t,
		err.Error(),
		fmt.Sprintf("Ссылка %q не найдена", linkURL),
		"DeleteLink should return error",
	)
}

func TestClient_GetLinks_Success(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)
	apiLinks := []api.LinkResponse{
		{
			ID:  api.NewOptInt64(1),
			URL: api.NewOptURI(url.URL{Scheme: "https", Host: "example.com"}),
		},
		{
			ID:  api.NewOptInt64(2),
			URL: api.NewOptURI(url.URL{Scheme: "https", Host: "example.org"}),
		},
	}

	clientMock.On("LinksGet", mock.Anything, api.LinksGetParams{TgChatID: chatID, Tag: api.NewOptString("tag")}).
		Return(&api.ListLinksResponse{Links: apiLinks}, nil).Once()

	links, err := client.GetLinks(context.Background(), chatID, "tag")

	assert.NoError(t, err, "GetLinks should not return error")
	assert.Len(t, links, 2, "GetLinks should return 2 links")
	assert.Equal(
		t,
		exampleLink,
		links[0].URL,
		"GetLinks should return link with correct URL",
	)
	assert.Equal(
		t,
		"https://example.org",
		links[1].URL,
		"GetLinks should return link with correct URL",
	)
}

func TestClient_GetLinks_Error(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)
	expectedErr := errors.New("database error")

	clientMock.On("LinksGet", mock.Anything, api.LinksGetParams{TgChatID: chatID}).
		Return(nil, expectedErr).Once()

	links, err := client.GetLinks(context.Background(), chatID, "")

	assert.Error(t, err, "GetLinks should return error")
	assert.Nil(t, links, "GetLinks should return empty list")
	assert.Contains(t, err.Error(), "failed to get links", "GetLinks should return error")
}

func TestClient_GetLinks_APIError(t *testing.T) {
	t.Parallel()

	clientMock := mocks.NewMockExternalClient(t)
	client := scrapper.NewClient(clientMock)

	chatID := int64(12345)

	clientMock.On("LinksGet", mock.Anything, api.LinksGetParams{TgChatID: chatID}).
		Return(&api.ApiErrorResponse{
			Description: api.NewOptString("error description"),
		}, nil).Once()

	links, err := client.GetLinks(context.Background(), chatID, "")

	errResp := scrapper.ErrResponse{}
	require.True(t, errors.As(err, &errResp), "RegisterChat should return error response")
	assert.Equal(
		t,
		"failed to get links: error description",
		errResp.Message,
		"RegisterChat should return error response",
	)
	assert.Nil(t, links, "GetLinks should return empty list")
}
