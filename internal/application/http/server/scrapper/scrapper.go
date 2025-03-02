package scrapper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	repository "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
)

type Repository interface {
	RegisterChat(chatID int64) error
	DeleteChat(chatID int64) error
	TrackLink(link *domain.Link) (*domain.Link, error)
	UntrackLink(chatID int64, url string) (*domain.Link, error)
	ListLinks(chatID int64) ([]*domain.Link, error)
}

type Server struct {
	repo Repository
}

func NewServer(repo Repository) *Server {
	return &Server{
		repo: repo,
	}
}

func (s *Server) TgChatIDPost(
	_ context.Context,
	params scrapper.TgChatIDPostParams,
) (scrapper.TgChatIDPostRes, error) {
	err := s.repo.RegisterChat(params.ID)
	if err != nil {
		return &scrapper.ApiErrorResponse{
			Code:        scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
			Description: scrapper.NewOptString(err.Error()),
		}, nil
	}

	return &scrapper.TgChatIDPostOK{}, nil
}

func (s *Server) TgChatIDDelete(
	_ context.Context,
	params scrapper.TgChatIDDeleteParams,
) (scrapper.TgChatIDDeleteRes, error) {
	err := s.repo.DeleteChat(params.ID)

	switch {
	case errors.As(err, &repository.ErrUnregister{}):
		return &scrapper.TgChatIDDeleteNotFound{}, nil

	case err != nil:
		return &scrapper.TgChatIDDeleteBadRequest{
			Code:        scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
			Description: scrapper.NewOptString(err.Error()),
		}, nil

	default:
		return &scrapper.TgChatIDDeleteOK{}, nil
	}
}

func (s *Server) LinksPost(
	_ context.Context,
	req *scrapper.AddLinkRequest,
	params scrapper.LinksPostParams,
) (scrapper.LinksPostRes, error) {
	link, err := s.repo.TrackLink(&domain.Link{
		ChatID:  params.TgChatID,
		URL:     req.Link.Value.String(),
		Tags:    req.Tags,
		Filters: req.Filters,
	})
	if err != nil {
		return &scrapper.ApiErrorResponse{
			Code:        scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
			Description: scrapper.NewOptString(err.Error()),
		}, nil
	}

	parsedURL, err := url.Parse(link.URL)
	if err != nil {
		return &scrapper.ApiErrorResponse{
			Code:        scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
			Description: scrapper.NewOptString(fmt.Sprintf("failed to parse url: %s", err)),
		}, nil
	}

	return &scrapper.LinkResponse{
		ID:      scrapper.NewOptInt64(link.ID),
		URL:     scrapper.NewOptURI(*parsedURL),
		Tags:    link.Tags,
		Filters: link.Filters,
	}, nil
}

func (s *Server) LinksGet(
	_ context.Context,
	params scrapper.LinksGetParams,
) (scrapper.LinksGetRes, error) {
	links, err := s.repo.ListLinks(params.TgChatID)
	if err != nil {
		return &scrapper.ApiErrorResponse{
			Code:        scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
			Description: scrapper.NewOptString(err.Error()),
		}, nil
	}

	return domainLinksToResponse(links), nil
}

func (s *Server) LinksDelete(
	_ context.Context,
	req *scrapper.RemoveLinkRequest,
	params scrapper.LinksDeleteParams,
) (scrapper.LinksDeleteRes, error) {
	link, err := s.repo.UntrackLink(params.TgChatID, req.Link.Value.String())

	switch {
	case errors.As(err, &repository.ErrUnregister{}):
		return &scrapper.LinksDeleteNotFound{}, nil

	case err != nil:
		return &scrapper.LinksDeleteBadRequest{
			Code:        scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
			Description: scrapper.NewOptString(err.Error()),
		}, nil

	default:
		parsedURL, err := url.Parse(link.URL)
		if err != nil {
			return &scrapper.LinksDeleteBadRequest{
				Code: scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
				Description: scrapper.NewOptString(
					fmt.Sprintf("failed to parse url=%q: %s", link.URL, err),
				),
			}, nil
		}

		return &scrapper.LinkResponse{
			ID:      scrapper.NewOptInt64(link.ID),
			URL:     scrapper.NewOptURI(*parsedURL),
			Tags:    link.Tags,
			Filters: link.Filters,
		}, nil
	}
}

func domainLinksToResponse(links []*domain.Link) scrapper.LinksGetRes {
	respLinks := make([]scrapper.LinkResponse, 0, len(links))

	for _, link := range links {
		parsedURL, err := url.Parse(link.URL)
		if err != nil {
			return &scrapper.ApiErrorResponse{
				Code: scrapper.NewOptString(http.StatusText(http.StatusInternalServerError)),
				Description: scrapper.NewOptString(
					fmt.Sprintf("failed to parse url=%q: %s", link.URL, err),
				),
			}
		}

		respLinks = append(respLinks, scrapper.LinkResponse{
			ID:      scrapper.NewOptInt64(link.ID),
			URL:     scrapper.NewOptURI(*parsedURL),
			Tags:    link.Tags,
			Filters: link.Filters,
		})
	}

	//nolint:gosec //has overflow check
	var size = int32(min(len(respLinks), math.MaxInt32))

	return &scrapper.ListLinksResponse{
		Links: respLinks,
		Size:  scrapper.NewOptInt32(size),
	}
}
