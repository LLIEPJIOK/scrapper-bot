package bot

import (
	"context"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Repository interface {
	AddUpdate(ctx context.Context, update *domain.Update) error
}

type Server struct {
	repo     Repository
	channels *domain.Channels
}

func NewServer(repo Repository, channels *domain.Channels) *Server {
	return &Server{
		repo:     repo,
		channels: channels,
	}
}

func (s *Server) UpdatesPost(
	ctx context.Context,
	req *bot.LinkUpdate,
) (bot.UpdatesPostRes, error) {
	if req.GetSendImmediately().Value {
		msg := tgbotapi.NewMessage(req.GetChatID().Value, req.GetMessage().Value)
		msg.ParseMode = tgbotapi.ModeHTML

		s.channels.TelegramResp() <- msg

		return &bot.UpdatesPostOK{}, nil
	}

	err := s.repo.AddUpdate(ctx, &domain.Update{
		ChatID:  req.GetChatID().Value,
		URL:     req.URL.Value.String(),
		Message: req.GetMessage().Value,
		Tags:    req.GetTags(),
	})
	if err != nil {
		return &bot.ApiErrorResponse{
			Code:        bot.NewOptString(http.StatusText(http.StatusInternalServerError)),
			Description: bot.NewOptString(err.Error()),
		}, err
	}

	return &bot.UpdatesPostOK{}, nil
}
