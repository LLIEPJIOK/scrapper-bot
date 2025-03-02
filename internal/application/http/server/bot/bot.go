package bot

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

type BotServer struct {
	repo Repository
}

func NewBotServer(repo Repository) *BotServer {
	return &BotServer{repo: repo}
}

func (s *BotServer) UpdatesPost(
	_ context.Context,
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
