package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

const (
	prefix      = "https://stackoverflow.com/questions/"
	answersURL  = "https://api.stackexchange.com/2.3/questions/%s/answers?fromdate=%d&&site=stackoverflow"
	commentsURL = "https://api.stackexchange.com/2.3/questions/%s/comments?fromdate=%d&site=stackoverflow"
)

type Client struct {
	client *http.Client
}

func New(httpClient *http.Client) *Client {
	return &Client{
		client: httpClient,
	}
}

func (c *Client) HasUpdates(link string, lastUpdate int64) (bool, error) {
	suffix, ok := strings.CutPrefix(link, prefix)
	if !ok {
		return false, NewErrInvalidLink(link)
	}

	slashIdx := strings.Index(suffix, "/")
	if slashIdx == -1 {
		return false, NewErrInvalidLink(link)
	}

	questionID := suffix[:slashIdx]

	hasUpdates, err := c.hasSourceUpdates(answersURL, questionID, lastUpdate)
	if err != nil {
		return false, fmt.Errorf(
			"c.hasSourceUpdates(%q, %q, %d): %w",
			answersURL,
			questionID,
			lastUpdate,
			err,
		)
	}

	if hasUpdates {
		return true, nil
	}

	hasUpdates, err = c.hasSourceUpdates(commentsURL, questionID, lastUpdate)
	if err != nil {
		return false, fmt.Errorf(
			"c.hasSourceUpdates(%q, %q, %d): %w",
			commentsURL,
			questionID,
			lastUpdate,
			err,
		)
	}

	return hasUpdates, nil
}

func (c *Client) hasSourceUpdates(sourceURL, questionID string, lastUpdate int64) (bool, error) {
	url := fmt.Sprintf(sourceURL, questionID, lastUpdate)

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

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
	data := &Data{}

	if err := dec.Decode(data); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return len(data.Items) != 0, nil
}
