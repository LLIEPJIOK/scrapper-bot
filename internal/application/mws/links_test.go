package mws_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/mws"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/mws/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewLinksCounter(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		requestBody    any
		responseStatus int
		setupMocks     func(mockMetrics *mocks.MockMetrics)
	}{
		{
			name:   "POST GitHub link - success",
			method: http.MethodPost,
			path:   "/links",
			requestBody: map[string]string{
				"link": "https://github.com/user/repo",
			},
			responseStatus: http.StatusOK,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("IncActiveLinksTotal", "github").Once()
			},
		},
		{
			name:   "POST StackOverflow link - success",
			method: http.MethodPost,
			path:   "/links",
			requestBody: map[string]string{
				"link": "https://stackoverflow.com/questions/123",
			},
			responseStatus: http.StatusOK,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("IncActiveLinksTotal", "stack_overflow").Once()
			},
		},
		{
			name:   "POST unknown link - success",
			method: http.MethodPost,
			path:   "/links",
			requestBody: map[string]string{
				"link": "https://example.com/some-link",
			},
			responseStatus: http.StatusOK,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("IncActiveLinksTotal", "unknown").Once()
			},
		},
		{
			name:   "DELETE GitHub link - success",
			method: http.MethodDelete,
			path:   "/links",
			requestBody: map[string]string{
				"link": "https://github.com/user/repo",
			},
			responseStatus: http.StatusOK,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("DecActiveLinksTotal", "github").Once()
			},
		},
		{
			name:   "POST link - non-200 response",
			method: http.MethodPost,
			path:   "/links",
			requestBody: map[string]string{
				"link": "https://github.com/user/repo",
			},
			responseStatus: http.StatusBadRequest,
			setupMocks: func(_ *mocks.MockMetrics) {
			},
		},
		{
			name:   "GET request - should skip",
			method: http.MethodGet,
			path:   "/links",
			requestBody: map[string]string{
				"link": "https://github.com/user/repo",
			},
			responseStatus: http.StatusOK,
			setupMocks: func(_ *mocks.MockMetrics) {
			},
		},
		{
			name:   "POST different path - should skip",
			method: http.MethodPost,
			path:   "/other",
			requestBody: map[string]string{
				"link": "https://github.com/user/repo",
			},
			responseStatus: http.StatusOK,
			setupMocks: func(_ *mocks.MockMetrics) {
			},
		},
		{
			name:           "POST invalid JSON - should skip metrics",
			method:         http.MethodPost,
			path:           "/links",
			requestBody:    "invalid json",
			responseStatus: http.StatusOK,
			setupMocks: func(_ *mocks.MockMetrics) {
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMetrics := mocks.NewMockMetrics(t)
			tc.setupMocks(mockMetrics)

			middleware := mws.NewLinksCounter(mockMetrics)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.responseStatus)
				fmt.Fprint(w, "OK")
			})

			wrappedHandler := middleware(nextHandler)

			requestBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err, "Failed to marshal request body")

			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			assert.Equal(t, tc.responseStatus, rr.Code, "Response status code")
		})
	}
}
