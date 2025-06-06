package scrapper_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/server/http/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/server/http/scrapper/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	repository "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/scrapper"
	api "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validURL = "http://example.com"

func TestTgChatIDPost_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repoMock := mocks.NewMockRepository(t)
	repoMock.On("RegisterChat", ctx, int64(123)).Return(nil).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.TgChatIDPostParams{ID: 123}
	res, err := srv.TgChatIDPost(ctx, params)
	require.NoError(t, err, "Expected no error on successful registration")

	_, ok := res.(*api.TgChatIDPostOK)
	assert.True(t, ok, "Expected response to be TgChatIDPostOK")
}

func TestTgChatIDPost_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expectedErr := errors.New("registration error")
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("RegisterChat", ctx, int64(456)).Return(expectedErr).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.TgChatIDPostParams{ID: 456}

	res, err := srv.TgChatIDPost(ctx, params)
	require.NoError(t, err, "Expected no transport error")

	apiErr, ok := res.(*api.ApiErrorResponse)
	require.True(t, ok, "Expected response to be ApiErrorResponse")
	assert.Equal(
		t,
		http.StatusText(http.StatusInternalServerError),
		apiErr.Code.Value,
		"Expected proper error code",
	)
	assert.Equal(
		t,
		expectedErr.Error(),
		apiErr.Description.Value,
		"Expected error description to match",
	)
}

func TestTgChatIDDelete_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repoMock := mocks.NewMockRepository(t)
	repoMock.On("DeleteChat", ctx, int64(789)).Return(nil).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.TgChatIDDeleteParams{ID: 789}
	res, err := srv.TgChatIDDelete(ctx, params)
	require.NoError(t, err, "Expected no error on successful delete")

	_, ok := res.(*api.TgChatIDDeleteOK)
	assert.True(t, ok, "Expected response to be TgChatIDDeleteOK")
}

func TestTgChatIDDelete_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	unregErr := repository.ErrUnregister{}
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("DeleteChat", ctx, int64(101)).Return(unregErr).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.TgChatIDDeleteParams{ID: 101}

	res, err := srv.TgChatIDDelete(ctx, params)
	require.NoError(t, err, "Expected no transport error")

	_, ok := res.(*api.TgChatIDDeleteNotFound)
	assert.True(t, ok, "Expected response to be TgChatIDDeleteNotFound")
}

func TestTgChatIDDelete_GenericError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	genErr := errors.New("delete error")
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("DeleteChat", ctx, int64(202)).Return(genErr).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.TgChatIDDeleteParams{ID: 202}
	res, err := srv.TgChatIDDelete(ctx, params)
	require.NoError(t, err, "Expected no transport error")

	badReq, ok := res.(*api.TgChatIDDeleteBadRequest)
	require.True(t, ok, "Expected response to be TgChatIDDeleteBadRequest")
	assert.Equal(
		t,
		http.StatusText(http.StatusInternalServerError),
		badReq.Code.Value,
		"Expected error code to match",
	)
	assert.Equal(t, genErr.Error(), badReq.Description.Value, "Expected error description to match")
}

func TestLinksPost_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	parsedValidURL, err := url.Parse(validURL)
	require.NoError(t, err, "Expected no error on valid URL")

	repoMock := mocks.NewMockRepository(t)
	repoMock.On("TrackLink", ctx, mock.MatchedBy(func(link *domain.Link) bool {
		return link.ChatID == 333 &&
			link.URL == validURL &&
			len(link.Tags) == 2 && link.Tags[0] == "tag1" && link.Tags[1] == "tag2" &&
			len(link.Filters) == 1 && link.Filters[0] == "filter1"
	})).Return(&domain.Link{
		ID:              1001,
		ChatID:          333,
		URL:             validURL,
		Tags:            []string{"tag1", "tag2"},
		Filters:         []string{"filter1"},
		SendImmediately: domain.NewNull(true),
	}, nil).Once()

	srv := scrapper.NewServer(repoMock)

	req := &api.AddLinkRequest{
		Link:            api.NewOptURI(*parsedValidURL),
		Tags:            []string{"tag1", "tag2"},
		Filters:         []string{"filter1"},
		SendImmediately: api.NewOptBool(true),
	}
	params := api.LinksPostParams{TgChatID: 333}

	res, err := srv.LinksPost(ctx, req, params)
	require.NoError(t, err, "Expected no error on successful link tracking")

	linkResp, ok := res.(*api.LinkResponse)
	require.True(t, ok, "Expected response to be LinkResponse")
	assert.Equal(t, int64(1001), linkResp.ID.Value, "Expected link ID to match")

	parsedURL, err := url.Parse(validURL)
	require.NoError(t, err, "Expected URL to parse successfully")
	assert.Equal(t, parsedURL.String(), linkResp.URL.Value.String(), "Expected URL to match")
	assert.Equal(t, req.Tags, linkResp.Tags, "Expected tags to match")
	assert.Equal(t, req.Filters, linkResp.Filters, "Expected filters to match")
}

func TestLinksPost_TrackLinkError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	parsedValidURL, err := url.Parse(validURL)
	require.NoError(t, err, "Expected no error on valid URL")

	expectedErr := errors.New("track error")
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("TrackLink", ctx, mock.Anything).Return(nil, expectedErr).Once()

	srv := scrapper.NewServer(repoMock)

	req := &api.AddLinkRequest{
		Link:            api.NewOptURI(*parsedValidURL),
		Tags:            []string{},
		Filters:         []string{},
		SendImmediately: api.NewOptBool(false),
	}
	params := api.LinksPostParams{TgChatID: 444}

	res, err := srv.LinksPost(ctx, req, params)
	require.NoError(t, err, "Expected no transport error")

	apiErr, ok := res.(*api.ApiErrorResponse)
	require.True(t, ok, "Expected response to be ApiErrorResponse")
	assert.Equal(
		t,
		http.StatusText(http.StatusInternalServerError),
		apiErr.Code.Value,
		"Expected error code to match",
	)
	assert.Equal(
		t,
		expectedErr.Error(),
		apiErr.Description.Value,
		"Expected error description to match",
	)
}

func TestLinksGet_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	links := []*domain.Link{
		{
			ID:              1,
			URL:             validURL,
			Tags:            []string{"a"},
			Filters:         []string{"x"},
			SendImmediately: domain.NewNull(true),
		},
		{
			ID:              2,
			URL:             validURL,
			Tags:            []string{"b"},
			Filters:         []string{"y"},
			SendImmediately: domain.NewNull(false),
		},
	}
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("ListLinks", ctx, int64(777)).Return(links, nil).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.LinksGetParams{TgChatID: 777}
	res, err := srv.LinksGet(ctx, params)
	require.NoError(t, err, "Expected no error on successful listing")

	listResp, ok := res.(*api.ListLinksResponse)
	require.True(t, ok, "Expected response to be ListLinksResponse")

	expectedCount := 2
	assert.Len(t, listResp.Links, expectedCount, "Expected correct number of links")
	assert.Equal(
		t,
		int32(expectedCount),
		listResp.Size.Value,
		"Expected size to match number of links",
	)

	repoMock.AssertExpectations(t)
}

func TestLinksGetBy_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	links := []*domain.Link{
		{ID: 1, URL: validURL, Tags: []string{"a"}, Filters: []string{"x"}},
		{ID: 2, URL: validURL, Tags: []string{"b"}, Filters: []string{"y"}},
	}
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("ListLinksByTag", ctx, int64(777), "tag").Return(links, nil).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.LinksGetParams{TgChatID: 777, Tag: api.NewOptString("tag")}
	res, err := srv.LinksGet(ctx, params)
	require.NoError(t, err, "Expected no error on successful listing")

	listResp, ok := res.(*api.ListLinksResponse)
	require.True(t, ok, "Expected response to be ListLinksResponse")

	expectedCount := 2
	assert.Len(t, listResp.Links, expectedCount, "Expected correct number of links")
	assert.Equal(
		t,
		int32(expectedCount),
		listResp.Size.Value,
		"Expected size to match number of links",
	)

	repoMock.AssertExpectations(t)
}

func TestLinksGet_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expectedErr := errors.New("list error")
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("ListLinks", ctx, int64(888)).Return(nil, expectedErr).Once()

	srv := scrapper.NewServer(repoMock)
	params := api.LinksGetParams{TgChatID: 888}

	res, err := srv.LinksGet(ctx, params)
	require.NoError(t, err, "Expected no transport error")

	apiErr, ok := res.(*api.ApiErrorResponse)
	require.True(t, ok, "Expected response to be ApiErrorResponse")

	expectedCode := http.StatusText(http.StatusInternalServerError)
	assert.Equal(t, expectedCode, apiErr.Code.Value, "Expected error code to match")
	assert.Equal(
		t,
		expectedErr.Error(),
		apiErr.Description.Value,
		"Expected error description to match",
	)
}

func TestLinksDelete_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	parsedValidURL, err := url.Parse(validURL)
	require.NoError(t, err, "Expected no error on valid URL")

	unregErr := repository.ErrUnregister{}
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("UntrackLink", ctx, int64(111), mock.Anything).Return(nil, unregErr).Once()

	srv := scrapper.NewServer(repoMock)

	req := &api.RemoveLinkRequest{
		Link: api.NewOptURI(*parsedValidURL),
	}
	params := api.LinksDeleteParams{TgChatID: 111}

	res, err := srv.LinksDelete(ctx, req, params)
	require.NoError(t, err, "Expected no transport error")

	_, ok := res.(*api.LinksDeleteNotFound)
	assert.True(t, ok, "Expected response to be LinksDeleteNotFound")
}

func TestLinksDelete_GenericError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	parsedValidURL, err := url.Parse(validURL)
	require.NoError(t, err, "Expected no error on valid URL")

	genErr := errors.New("delete error")
	repoMock := mocks.NewMockRepository(t)
	repoMock.On("UntrackLink", ctx, int64(222), mock.Anything).Return(nil, genErr).Once()

	srv := scrapper.NewServer(repoMock)

	req := &api.RemoveLinkRequest{
		Link: api.NewOptURI(*parsedValidURL),
	}
	params := api.LinksDeleteParams{TgChatID: 222}

	res, err := srv.LinksDelete(ctx, req, params)
	require.NoError(t, err, "Expected no transport error")

	badReq, ok := res.(*api.LinksDeleteBadRequest)
	require.True(t, ok, "Expected response to be LinksDeleteBadRequest")

	expectedCode := http.StatusText(http.StatusInternalServerError)
	assert.Equal(t, expectedCode, badReq.Code.Value, "Expected error code to match")
	assert.Equal(t, genErr.Error(), badReq.Description.Value, "Expected error description to match")
}

func TestLinksDelete_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	parsedValidURL, err := url.Parse(validURL)
	require.NoError(t, err, "Expected no error on valid URL")

	repoMock := mocks.NewMockRepository(t)
	repoMock.On("UntrackLink", ctx, int64(444), mock.Anything).Return(&domain.Link{
		ID:      444,
		URL:     validURL,
		Tags:    []string{"a"},
		Filters: []string{"b"},
	}, nil).Once()

	srv := scrapper.NewServer(repoMock)

	req := &api.RemoveLinkRequest{
		Link: api.NewOptURI(*parsedValidURL),
	}
	params := api.LinksDeleteParams{TgChatID: 444}

	res, err := srv.LinksDelete(ctx, req, params)
	require.NoError(t, err, "Expected no error on successful link untracking")

	linkResp, ok := res.(*api.LinkResponse)
	require.True(t, ok, "Expected response to be LinkResponse")
	assert.Equal(t, int64(444), linkResp.ID.Value, "Expected link ID to match")

	parsedURL, err := url.Parse(validURL)
	require.NoError(t, err, "Expected URL to parse successfully")
	assert.Equal(t, parsedURL.String(), linkResp.URL.Value.String(), "Expected URL to match")
	assert.Equal(t, []string{"a"}, linkResp.Tags, "Expected tags to match")
	assert.Equal(t, []string{"b"}, linkResp.Filters, "Expected filters to match")
}
