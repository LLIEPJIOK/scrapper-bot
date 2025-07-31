package metrics_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/metrics"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/metrics/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMetricsMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		handlerStatus  int
		expectedStatus int
		setupMocks     func(mockMetrics *mocks.MockMetrics)
	}{
		{
			name:           "GET request with 200 status",
			method:         "GET",
			path:           "/api/v1/test",
			handlerStatus:  http.StatusOK,
			expectedStatus: http.StatusOK,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("IncHTTPRequestsTotal", "GET", "/api/v1/test", http.StatusOK).Once()
				mockMetrics.On("ObserveHTTPRequestsDurationSeconds", "GET", "/api/v1/test", mock.AnythingOfType("float64")).
					Once()
			},
		},
		{
			name:           "POST request with 404 status",
			method:         "POST",
			path:           "/api/v1/notfound",
			handlerStatus:  http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("IncHTTPRequestsTotal", "POST", "/api/v1/notfound", http.StatusNotFound).
					Once()
				mockMetrics.On("ObserveHTTPRequestsDurationSeconds", "POST", "/api/v1/notfound", mock.AnythingOfType("float64")).
					Once()
			},
		},
		{
			name:           "PUT request with 500 status",
			method:         "PUT",
			path:           "/api/v1/error",
			handlerStatus:  http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("IncHTTPRequestsTotal", "PUT", "/api/v1/error", http.StatusInternalServerError).
					Once()
				mockMetrics.On("ObserveHTTPRequestsDurationSeconds", "PUT", "/api/v1/error", mock.AnythingOfType("float64")).
					Once()
			},
		},
		{
			name:           "DELETE request without explicit status (should default to 200)",
			method:         "DELETE",
			path:           "/api/v1/resource",
			handlerStatus:  0, // No explicit status set
			expectedStatus: http.StatusOK,
			setupMocks: func(mockMetrics *mocks.MockMetrics) {
				mockMetrics.On("IncHTTPRequestsTotal", "DELETE", "/api/v1/resource", http.StatusOK).
					Once()
				mockMetrics.On("ObserveHTTPRequestsDurationSeconds", "DELETE", "/api/v1/resource", mock.AnythingOfType("float64")).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMetrics := mocks.NewMockMetrics(t)
			tc.setupMocks(mockMetrics)

			middleware := metrics.New(mockMetrics)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.handlerStatus != 0 {
					w.WriteHeader(tc.handlerStatus)
				}

				fmt.Fprint(w, "test response")
			})

			wrappedHandler := middleware(testHandler)

			req := httptest.NewRequest(tc.method, "http://example.com"+tc.path, http.NoBody)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, "Expected status code does not match")
			assert.Contains(
				t,
				rr.Body.String(),
				"test response",
				"Response body does not match expected content",
			)
		})
	}
}

func TestMetricsMiddleware_DurationMeasurement(t *testing.T) {
	mockMetrics := mocks.NewMockMetrics(t)

	expectedMinDuration := 0.1

	mockMetrics.On("IncHTTPRequestsTotal", "GET", "/slow", http.StatusOK).Once()
	mockMetrics.On("ObserveHTTPRequestsDurationSeconds", "GET", "/slow", mock.MatchedBy(func(duration float64) bool {
		return duration >= expectedMinDuration
	})).
		Once()

	middleware := metrics.New(mockMetrics)

	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "slow response")
	})

	wrappedHandler := middleware(slowHandler)

	req := httptest.NewRequest("GET", "http://example.com/slow", http.NoBody)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code does not match")
}
