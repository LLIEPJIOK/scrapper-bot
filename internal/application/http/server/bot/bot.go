package bot

import (
	"context"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
)

type Repository interface {
	AddUpdate(ctx context.Context, update *domain.Update) error
}

type Server struct {
	repo Repository
}

func NewServer(repo Repository) *Server {
	return &Server{repo: repo}
}

func (s *Server) UpdatesPost(
	ctx context.Context,
	req *bot.LinkUpdate,
) (bot.UpdatesPostRes, error) {
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
