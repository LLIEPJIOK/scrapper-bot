package health_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/server/http/health"
	"github.com/stretchr/testify/assert"
)

func TestServer_RegisterRoutes(t *testing.T) {
	server := health.New()
	mux := http.NewServeMux()

	server.RegisterRoutes(mux)

	t.Run("Health endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Health endpoint should return 200 OK")
		assert.Equal(t, "healthy", rr.Body.String(), "Health endpoint should return 'healthy'")
	})

	t.Run("Metrics endpoint", func(t *testing.T) {
		metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
		metricsRr := httptest.NewRecorder()

		mux.ServeHTTP(metricsRr, metricsReq)

		assert.Equal(t, http.StatusOK, metricsRr.Code, "Metrics endpoint should return 200 OK")

		responseBody := metricsRr.Body.String()
		assert.Contains(
			t,
			responseBody,
			"# HELP",
			"Metrics response should contain Prometheus help comments",
		)
	})
}
