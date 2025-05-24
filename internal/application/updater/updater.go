package updater

import (
	"context"
	"errors"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

type Handler interface {
	UpdatesPost(ctx context.Context, update *domain.Update) error
}

type Updater struct {
	handlers []Handler
}

func New(handlers ...Handler) *Updater {
	return &Updater{
		handlers: handlers,
	}
}

func (u *Updater) UpdatesPost(ctx context.Context, update *domain.Update) error {
	for i, handler := range u.handlers {
		err := handler.UpdatesPost(ctx, update)
		if errors.As(err, &client.ErrServiceUnavailable{}) {
			slog.Error("failed to send update", slog.Int("handler_num", i), slog.Any("error", err))

			continue
		}

		if err != nil {
			return err
		}

		return nil
	}

	return NewErrSendUpdate()
}
