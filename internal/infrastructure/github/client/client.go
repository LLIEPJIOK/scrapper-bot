package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

const (
	repoURL  = "https://api.github.com/repos/%s/%s"
	issueURL = "https://api.github.com/repos/%s/%s/issues/%s/timeline"
	pullURL  = "https://api.github.com/repos/%s/%s/pulls/%s"
)

type Client struct {
	client     *http.Client
	repoRegex  *regexp.Regexp
	issueRegex *regexp.Regexp
	pullRegex  *regexp.Regexp
	token      string
}

func New(cfg *config.GitHub, httpClient *http.Client) *Client {
	return &Client{
		client:     httpClient,
		repoRegex:  regexp.MustCompile(`^https://github\.com/([\w\.-]+)/([\w\.-]+)$`),
		issueRegex: regexp.MustCompile(`^https://github\.com/([\w\.-]+)/([\w\.-]+)/issues/(\d+)$`),
		pullRegex:  regexp.MustCompile(`^https://github\.com/([\w\.-]+)/([\w\.-]+)/pull/(\d+)$`),
		token:      cfg.Token,
	}
}

func (c *Client) HasUpdates(link string, lastCheck time.Time) (bool, error) {
	switch {
	case c.repoRegex.MatchString(link):
		matches := c.repoRegex.FindStringSubmatch(link)
		url := fmt.Sprintf(repoURL, matches[1], matches[2])

		return c.hasSourceUpdates(url, lastCheck)

	case c.issueRegex.MatchString(link):
		matches := c.issueRegex.FindStringSubmatch(link)
		url := fmt.Sprintf(issueURL, matches[1], matches[2], matches[3])

		return c.hasSourceUpdates(url, lastCheck)

	case c.pullRegex.MatchString(link):
		matches := c.pullRegex.FindStringSubmatch(link)
		url := fmt.Sprintf(pullURL, matches[1], matches[2], matches[3])

		return c.hasSourceUpdates(url, lastCheck)

	default:
		return false, NewErrInvalidLink(link, "unknown type")
	}
}

func (c *Client) hasSourceUpdates(url string, lastCheck time.Time) (bool, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", c.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to get response: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error(
				"failed to close response body",
				slog.Any("error", err),
				slog.Any("service", "stackoverflow client"),
				slog.Any("method", "hasSourceUpdates"),
				slog.Any("url", url),
			)
		}
	}()

	dec := json.NewDecoder(resp.Body)
	data := make([]Data, 0)

	if err := dec.Decode(&data); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, d := range data {
		if d.UpdatedAt.After(lastCheck) {
			return true, nil
		}
	}

	return false, nil
}
