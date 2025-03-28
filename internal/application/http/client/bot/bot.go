package bot

import (
	"context"
	"fmt"
	"net/url"

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

func (b *Client) UpdatesPost(ctx context.Context, link string, chats []int64) error {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("failed to parse link: %w", err)
	}

	rawResp, err := b.client.UpdatesPost(ctx, &bot.LinkUpdate{
		URL:       bot.NewOptURI(*parsedURL),
		TgChatIds: chats,
	})
	if err != nil {
		return fmt.Errorf("failed to send updates: %w", err)
	}

	switch resp := rawResp.(type) {
	case *bot.UpdatesPostOK:
		return nil

	case *bot.ApiErrorResponse:
		return NewErrResponse(fmt.Sprintf("failed to add link: %s", resp.Description.Value))

	default:
		return NewErrResponse("invalid response type")
	}
}
