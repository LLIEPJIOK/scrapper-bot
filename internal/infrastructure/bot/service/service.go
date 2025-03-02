package service

import (
	"context"
	"net/http"

	repository "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/bot"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
)

type Repository interface {
	AddLink(chatID int64, link string) error
	GetUpdates() ([]*repository.UpdateChat, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdatesPost(
	ctx context.Context,
	req *bot.LinkUpdate,
) (bot.UpdatesPostRes, error) {
	for _, chat := range req.TgChatIds {
		err := s.repo.AddLink(chat, req.URL.Value.String())
		if err != nil {
			return &bot.ApiErrorResponse{
				Code:        bot.NewOptString(http.StatusText(http.StatusInternalServerError)),
				Description: bot.NewOptString(err.Error()),
			}, err
		}
	}

	return &bot.UpdatesPostOK{}, nil
}
