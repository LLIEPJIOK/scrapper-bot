package scrapper

import (
	"context"
	"fmt"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
)

type ExternalClient interface {
	LinksDelete(
		ctx context.Context,
		request *scrapper.RemoveLinkRequest,
		params scrapper.LinksDeleteParams,
	) (scrapper.LinksDeleteRes, error)
	LinksGet(ctx context.Context, params scrapper.LinksGetParams) (scrapper.LinksGetRes, error)
	LinksPost(
		ctx context.Context,
		request *scrapper.AddLinkRequest,
		params scrapper.LinksPostParams,
	) (scrapper.LinksPostRes, error)
	TgChatIDDelete(
		ctx context.Context,
		params scrapper.TgChatIDDeleteParams,
	) (scrapper.TgChatIDDeleteRes, error)
	TgChatIDPost(
		ctx context.Context,
		params scrapper.TgChatIDPostParams,
	) (scrapper.TgChatIDPostRes, error)
}

type Client struct {
	client ExternalClient
}

func NewClient(client ExternalClient) *Client {
	return &Client{
		client: client,
	}
}

func (s *Client) RegisterChat(ctx context.Context, id int64) error {
	rawResp, err := s.client.TgChatIDPost(ctx, scrapper.TgChatIDPostParams{
		ID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to register chat: %w", err)
	}

	switch resp := rawResp.(type) {
	case *scrapper.TgChatIDPostOK:
		return nil

	case *scrapper.ApiErrorResponse:
		return NewErrResponse(fmt.Sprintf("failed to register chat: %s", resp.Description.Value))

	case *scrapper.TgChatIDPostTooManyRequests:
		return NewErrUserResponse("Слишком много запросов. Повторите, пожалуйста, через некоторое время")

	default:
		return NewErrResponse("invalid response type")
	}
}

func (s *Client) AddLink(ctx context.Context, link *domain.Link) error {
	parsedURL, err := url.Parse(link.URL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	rawResp, err := s.client.LinksPost(ctx, &scrapper.AddLinkRequest{
		Link:            scrapper.NewOptURI(*parsedURL),
		Tags:            link.Tags,
		Filters:         link.Filters,
		SendImmediately: scrapper.NewOptBool(link.SendImmediately.Value),
	}, scrapper.LinksPostParams{
		TgChatID: link.ChatID,
	})
	if err != nil {
		return fmt.Errorf("failed to add link: %w", err)
	}

	switch resp := rawResp.(type) {
	case *scrapper.LinkResponse:
		return nil

	case *scrapper.ApiErrorResponse:
		return NewErrResponse(fmt.Sprintf("failed to add link: %s", resp.Description.Value))

	case *scrapper.LinksPostTooManyRequests:
		return NewErrUserResponse("Слишком много запросов. Повторите, пожалуйста, через некоторое время")

	default:
		return NewErrResponse("invalid response type")
	}
}

func (s *Client) DeleteLink(ctx context.Context, chatID int64, linkURL string) error {
	parsedURL, err := url.Parse(linkURL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	rawResp, err := s.client.LinksDelete(ctx, &scrapper.RemoveLinkRequest{
		Link: scrapper.NewOptURI(*parsedURL),
	}, scrapper.LinksDeleteParams{
		TgChatID: chatID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	switch resp := rawResp.(type) {
	case *scrapper.LinkResponse:
		return nil

	case *scrapper.LinksDeleteBadRequest:
		return NewErrResponse(fmt.Sprintf("bad request: %s", resp.Description.Value))

	case *scrapper.LinksDeleteNotFound:
		return NewErrUserResponse(fmt.Sprintf("Ссылка %q не найдена", linkURL))

	case *scrapper.LinksDeleteTooManyRequests:
		return NewErrUserResponse("Слишком много запросов. Повторите, пожалуйста, через некоторое время")

	default:
		return NewErrResponse("invalid response type")
	}
}

func (s *Client) GetLinks(ctx context.Context, chatID int64, tag string) ([]*domain.Link, error) {
	rawResp, err := s.client.LinksGet(ctx, scrapper.LinksGetParams{
		TgChatID: chatID,
		Tag:      scrapper.OptString{Value: tag, Set: tag != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	switch resp := rawResp.(type) {
	case *scrapper.ListLinksResponse:
		return linksToDomainLinks(resp.Links), nil

	case *scrapper.ApiErrorResponse:
		return nil, NewErrResponse(fmt.Sprintf("failed to get links: %s", resp.Description.Value))

	case *scrapper.LinksGetTooManyRequests:
		return nil, NewErrUserResponse("Слишком много запросов. Повторите, пожалуйста, через некоторое время")

	default:
		return nil, NewErrResponse("invalid response type")
	}
}

func linksToDomainLinks(links []scrapper.LinkResponse) []*domain.Link {
	domainLinks := make([]*domain.Link, 0, len(links))
	for i := range links {
		domainLinks = append(domainLinks, &domain.Link{
			ID:              links[i].ID.Value,
			URL:             links[i].URL.Value.String(),
			Tags:            links[i].Tags,
			Filters:         links[i].Filters,
			SendImmediately: domain.NewNull(links[i].SendImmediately.Value),
		})
	}

	return domainLinks
}
