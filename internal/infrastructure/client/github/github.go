package github

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

const (
	// repoActivityURL = "https://api.github.com/repos/%s/%s/activity"
	repoIssuesURL = "https://api.github.com/repos/%s/%s/issues"
	// repoPullsURL  = "https://api.github.com/repos/%s/%s/pulls"
	// repoBranchesURL = "https://api.github.com/repos/%s/%s/branches"

	// issueURL = "https://api.github.com/repos/%s/%s/issues/%s"
	// pullURL  = "https://api.github.com/repos/%s/%s/pulls/%s"
)

type GitHub struct {
	client    *http.Client
	repoRegex *regexp.Regexp
	// issueRegex *regexp.Regexp
	// pullRegex  *regexp.Regexp
	token    string
	pageSize string
}

func New(cfg *config.GitHub, httpClient *http.Client) *GitHub {
	return &GitHub{
		client:    httpClient,
		repoRegex: regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)$`),
		// issueRegex: regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)/issues/(\d+)$`),
		// pullRegex:  regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)/pull/(\d+)$`),
		token:    cfg.Token,
		pageSize: cfg.PageSize,
	}
}

func (g *GitHub) GetUpdates(link string, from, to time.Time) ([]string, error) {
	if !g.repoRegex.MatchString(link) {
		return []string{}, nil
	}

	matches := g.repoRegex.FindStringSubmatch(link)
	url := fmt.Sprintf(repoIssuesURL, matches[1], matches[2])

	msgs, err := g.getMessages(url, from, to)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (g *GitHub) getMessages(url string, from, to time.Time) ([]string, error) {
	msgs := make([]string, 0)

	page := 1

	for {
		data := make([]Data, 0)

		err := g.getAndDecodeResponse(url, page, &data)
		if err != nil {
			return nil, err
		}

		if len(data) == 0 || from.After(data[0].CreatedAt) {
			return msgs, nil
		}

		for _, d := range data {
			if from.After(d.CreatedAt) {
				return msgs, nil
			}

			if to.Before(d.CreatedAt) {
				continue
			}

			msgs = append(msgs, DataToMessage(&d))
		}

		page++
	}
}

func (g *GitHub) getAndDecodeResponse(link string, page int, data any) error {
	params := url.Values{}
	params.Add("sort", "created")
	params.Add("direction", "desc")
	params.Add("per_page", g.pageSize)
	params.Add("page", strconv.Itoa(page))

	reqURL := fmt.Sprintf("%s?%s", link, params.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request with url=%q: %w", reqURL, err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", g.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get response with url=%q: %w", reqURL, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error(
				"failed to close response body",
				slog.Any("error", err),
				slog.Any("service", "stackoverflow client"),
				slog.Any("method", "hasSourceUpdates"),
				slog.Any("url", reqURL),
			)
		}
	}()

	dec := json.NewDecoder(resp.Body)

	if err := dec.Decode(data); err != nil {
		return fmt.Errorf("failed to decode response with url=%q: %w", reqURL, err)
	}

	return nil
}
