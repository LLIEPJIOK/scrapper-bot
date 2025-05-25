package sof

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

const (
	prefix             = "https://stackoverflow.com/questions/"
	questionURL        = "https://api.stackexchange.com/2.3/questions/%s"
	answersURL         = "https://api.stackexchange.com/2.3/questions/%s/answers"
	answersCommentsURL = "https://api.stackexchange.com/2.3/answers/%s/comments"
	commentsURL        = "https://api.stackexchange.com/2.3/questions/%s/comments"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type SOF struct {
	client   Client
	pageSize string
}

func New(cfg *config.SOF, client Client) *SOF {
	return &SOF{
		client:   client,
		pageSize: cfg.PageSize,
	}
}

func (s *SOF) GetType() string {
	return "stack_overflow"
}

func (s *SOF) GetUpdates(link string, from, to time.Time) ([]string, error) {
	suffix, ok := strings.CutPrefix(link, prefix)
	if !ok {
		return []string{}, nil
	}

	slashIdx := strings.Index(suffix, "/")
	if slashIdx == -1 {
		return nil, NewErrInvalidLink(link)
	}

	questionID := suffix[:slashIdx]

	title, err := s.getTitle(questionID)
	if err != nil {
		return nil, err
	}

	answers, err := s.getAnswers(questionID, from, to)
	if err != nil {
		return nil, err
	}

	answersCommens, err := s.getAnswersComments(getAnswersIDs(answers), from, to)
	if err != nil {
		return nil, err
	}

	comments, err := s.getQuestionComments(questionID, from, to)
	if err != nil {
		return nil, err
	}

	msgs := make([]string, 0, len(answers)+len(answersCommens)+len(comments))
	msgs = append(msgs, AnswersToMessages(answers, title, link)...)
	msgs = append(msgs, CommentsToMessages(answersCommens, title, link)...)
	msgs = append(msgs, CommentsToMessages(comments, title, link)...)

	return msgs, nil
}

func (s *SOF) getTitle(questionID string) (string, error) {
	params := url.Values{}
	params.Add("site", "stackoverflow")

	reqURL := fmt.Sprintf(questionURL, questionID) + "?" + params.Encode()

	var data QuestionData

	if err := s.doRequest(reqURL, &data); err != nil {
		return "", err
	}

	if len(data.Items) == 0 {
		return "", NewErrQuestionNotFound(questionID)
	}

	return data.Items[0].Title, nil
}

func (s *SOF) getAnswers(questionID string, from, to time.Time) ([]Answer, error) {
	params := url.Values{}
	params.Add("site", "stackoverflow")
	params.Add("fromdate", fmt.Sprintf("%d", from.Unix()))
	params.Add("todate", fmt.Sprintf("%d", to.Unix()))
	params.Add("sort", "creation")
	params.Add("order", "desc")
	params.Add("filter", "withbody")

	page := 1
	answers := make([]Answer, 0)

	for {
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("pagesize", s.pageSize)

		reqURL := fmt.Sprintf(answersURL, questionID) + "?" + params.Encode()

		var data AnswerData

		if err := s.doRequest(reqURL, &data); err != nil {
			return nil, err
		}

		if len(data.Items) == 0 {
			return answers, nil
		}

		answers = append(answers, data.Items...)
		page++
	}
}

func (s *SOF) getAnswersComments(
	answersIDs []string,
	from, to time.Time,
) ([]Comment, error) {
	return s.getComments(fmt.Sprintf(answersCommentsURL, strings.Join(answersIDs, ";")), from, to)
}

func (s *SOF) getQuestionComments(questionID string, from, to time.Time) ([]Comment, error) {
	return s.getComments(fmt.Sprintf(commentsURL, questionID), from, to)
}

func (s *SOF) getComments(link string, from, to time.Time) ([]Comment, error) {
	params := url.Values{}
	params.Add("site", "stackoverflow")
	params.Add("fromdate", fmt.Sprintf("%d", from.Unix()))
	params.Add("todate", fmt.Sprintf("%d", to.Unix()))
	params.Add("sort", "creation")
	params.Add("order", "desc")
	params.Add("filter", "withbody")

	page := 1
	comments := make([]Comment, 0)

	for {
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("pagesize", s.pageSize)

		reqURL := link + "?" + params.Encode()

		var data CommentData

		if err := s.doRequest(reqURL, &data); err != nil {
			return nil, err
		}

		if len(data.Items) == 0 {
			return comments, nil
		}

		comments = append(comments, data.Items...)
		page++
	}
}

func (s *SOF) doRequest(reqURL string, data any) error {
	req, err := http.NewRequest(http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request with url=%q: %w", reqURL, err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get response with url=%q: %w", reqURL, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error(
				"failed to close response body",
				slog.Any("error", err),
				slog.Any("service", "stackoverflow client"),
				slog.Any("url", reqURL),
			)
		}
	}()

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return fmt.Errorf("failed to decode response with url=%q: %w", reqURL, err)
	}

	return nil
}

func getAnswersIDs(answers []Answer) []string {
	ids := make([]string, 0, len(answers))

	for _, answer := range answers {
		ids = append(ids, fmt.Sprintf("%d", answer.ID))
	}

	return ids
}
