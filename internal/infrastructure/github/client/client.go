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
	repoActivityURL = "https://api.github.com/repos/%s/%s/activity"
	repoIssuesURL   = "https://api.github.com/repos/%s/%s/issues"
	repoPullsURL    = "https://api.github.com/repos/%s/%s/pulls"
	repoBranchesURL = "https://api.github.com/repos/%s/%s/branches"

	issueURL = "https://api.github.com/repos/%s/%s/issues/%s"
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
		repoRegex:  regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)$`),
		issueRegex: regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)/issues/(\d+)$`),
		pullRegex:  regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)/pull/(\d+)$`),
		token:      cfg.Token,
	}
}

func (c *Client) HasUpdates(link string, lastCheck time.Time) (bool, error) {
	switch {
	case c.repoRegex.MatchString(link):
		matches := c.repoRegex.FindStringSubmatch(link)
		templates := []string{repoIssuesURL, repoPullsURL, repoBranchesURL}

		for _, template := range templates {
			url := fmt.Sprintf(template, matches[1], matches[2])
			data := make([]Data, 0)

			err := c.getAndDecodeResponse(url, &data)
			if err != nil {
				return false, err
			}

			for _, d := range data {
				if d.UpdatedAt.After(lastCheck) {
					return true, nil
				}
			}
		}

		url := fmt.Sprintf(repoActivityURL, matches[1], matches[2])
		data := make([]Data, 0)

		err := c.getAndDecodeResponse(url, &data)
		if err != nil {
			return false, err
		}

		for _, d := range data {
			if d.Timestamp.After(lastCheck) {
				return true, nil
			}
		}

		return false, nil

	case c.issueRegex.MatchString(link):
		matches := c.issueRegex.FindStringSubmatch(link)
		url := fmt.Sprintf(issueURL, matches[1], matches[2], matches[3])
		data := Data{}

		err := c.getAndDecodeResponse(url, &data)
		if err != nil {
			return false, err
		}

		return data.UpdatedAt.After(lastCheck), nil

	case c.pullRegex.MatchString(link):
		matches := c.pullRegex.FindStringSubmatch(link)
		url := fmt.Sprintf(pullURL, matches[1], matches[2], matches[3])
		data := Data{}

		err := c.getAndDecodeResponse(url, &data)
		if err != nil {
			return false, err
		}

		return data.UpdatedAt.After(lastCheck), nil

	default:
		return false, nil
	}
}

func (c *Client) getAndDecodeResponse(url string, data any) error {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request with url=%q: %w", url, err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", c.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get response with url=%q: %w", url, err)
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

	if err := dec.Decode(data); err != nil {
		return fmt.Errorf("failed to decode response with url=%q: %w", url, err)
	}

	return nil
}
