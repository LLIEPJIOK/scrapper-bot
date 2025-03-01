package app

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/scrapper/service"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
)

type runService = func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup)

func (a *App) services() []runService {
	return []runService{
		a.runServer,
	}
}

func (a *App) runServer(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("service stopped")

	svc := service.New(repository.New())

	srv, err := scrapper.NewServer(svc)
	if err != nil {
		slog.Error("failed to create scrapper server", slog.Any("error", err))

		return
	}

	if err := http.ListenAndServe(":8080", srv); err != nil {
		slog.Error("failed to start scrapper server", slog.Any("error", err))

		return
	}
}
