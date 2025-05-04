package bot

import (
	"context"
	"fmt"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
)

type ExternalClient interface {
	UpdatesPost(ctx context.Context, request *bot.LinkUpdate) (bot.UpdatesPostRes, error)
}

type Client struct {
	client ExternalClient
}

func NewClient(client ExternalClient) *Client {
	return &Client{
		client: client,
	}
}

func (b *Client) UpdatesPost(ctx context.Context, update *domain.Update) error {
	parsedURL, err := url.Parse(update.URL)
	if err != nil {
		return fmt.Errorf("failed to parse link: %w", err)
	}

	rawResp, err := b.client.UpdatesPost(ctx, &bot.LinkUpdate{
		ChatID:          bot.NewOptInt64(update.ChatID),
		URL:             bot.NewOptURI(*parsedURL),
		Message:         bot.NewOptString(update.Message),
		Tags:            update.Tags,
		SendImmediately: bot.NewOptBool(update.SendImmediately.Value),
	})
	if err != nil {
		return fmt.Errorf("failed to send updates: %w", err)
	}

	switch resp := rawResp.(type) {
	case *bot.UpdatesPostOK:
		return nil

	case *bot.ApiErrorResponse:
		return NewErrResponse(fmt.Sprintf("failed to add link: %s", resp.Description.Value))

	case *bot.UpdatesPostTooManyRequests:
		return NewErrResponse("too many requests")

	default:
		return NewErrResponse("invalid response type")
	}
}
