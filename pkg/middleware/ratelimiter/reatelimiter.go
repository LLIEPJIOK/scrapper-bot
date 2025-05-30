package ratelimiter

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/ratelimiter/repository"
)

func NewSlidingWindow(
	repo repository.Repository,
	cfg *Config,
) middleware.Middleware {
	window := cfg.Window
	maxHits := cfg.MaxHits

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			key := getKey(r, cfg.Name)
			now := time.Now()
			startWindow := now.Add(-window)

			err := repo.RemoveOldRecords(ctx, key, time.Time{}, startWindow)
			if err != nil {
				internalServerError(w, err)

				return
			}

			count, err := repo.CountRecords(ctx, key)
			if err != nil {
				internalServerError(w, err)

				return
			}

			if count >= int64(maxHits) {
				http.Error(
					w,
					http.StatusText(http.StatusTooManyRequests),
					http.StatusTooManyRequests,
				)

				return
			}

			err = repo.AddRecord(ctx, key, now)
			if err != nil {
				internalServerError(w, err)
				return
			}

			err = repo.ExpireKey(ctx, key, window)
			if err != nil {
				internalServerError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func getKey(r *http.Request, name string) string {
	ip := getClientIP(r)

	if name == "" {
		return fmt.Sprintf("ratelimiter:%s:%s", r.URL.Path, ip)
	}

	return fmt.Sprintf("ratelimiter:%s:%s:%s", r.URL.Path, name, ip)
}

func internalServerError(w http.ResponseWriter, err error) {
	slog.Error("internal server error", slog.Any("err", err))

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
