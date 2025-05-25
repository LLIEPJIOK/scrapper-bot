package metrics

import (
	"net/http"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware"
)

type Metrics interface {
	IncHTTPRequestsTotal(method, path string, status int)
	ObserveHTTPRequestsDurationSeconds(method, path string, seconds float64)
}

type customWriter struct {
	http.ResponseWriter
	status int
}

func (r *customWriter) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func New(m Metrics) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &customWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			start := time.Now()

			next.ServeHTTP(rec, r)

			dur := time.Since(start).Seconds()
			m.ObserveHTTPRequestsDurationSeconds(r.Method, r.URL.Path, dur)
			m.IncHTTPRequestsTotal(r.Method, r.URL.Path, rec.status)
		})
	}
}
