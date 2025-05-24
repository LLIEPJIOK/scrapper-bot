package mws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const (
	github        = "github"
	stackOverflow = "stack_overflow"
	unknown       = "unknown"
)

type Metrics interface {
	IncActiveLinksTotal(linkType string)
	DecActiveLinksTotal(linkType string)
}

type customWriter struct {
	http.ResponseWriter
	status int
}

func NewLinksCounter(m Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/links" ||
				(r.Method != http.MethodPost && r.Method != http.MethodDelete) {
				next.ServeHTTP(w, r)

				return
			}

			content, reader, err := readAndReturnReader(r.Body)
			if err != nil {
				slog.Error(
					"failed to read request body from links middleware",
					slog.Any("error", err),
				)
				next.ServeHTTP(w, r)

				return
			}

			r.Body = reader
			rec := &customWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			next.ServeHTTP(rec, r)

			if rec.status != http.StatusOK {
				return
			}

			var payload struct {
				Link string `json:"link"`
			}

			if err := json.Unmarshal(content, &payload); err != nil {
				slog.Error("failed to unmarshal link payload", slog.Any("error", err))

				return
			}

			linkType := getLinkType(payload.Link)

			switch r.Method {
			case http.MethodPost:
				m.IncActiveLinksTotal(linkType)

			case http.MethodDelete:
				m.DecActiveLinksTotal(linkType)
			}
		})
	}
}

func readAndReturnReader(r io.ReadCloser) ([]byte, io.ReadCloser, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read: %w", err)
	}

	err = r.Close()
	if err != nil {
		slog.Error("failed to close request body", slog.Any("error", err))
	}

	return content, io.NopCloser(bytes.NewReader(content)), nil
}

func getLinkType(link string) string {
	switch {
	case strings.HasPrefix(link, "https://github.com/"):
		return github

	case strings.HasPrefix(link, "https://stackoverflow.com/"):
		return stackOverflow

	default:
		return unknown
	}
}
