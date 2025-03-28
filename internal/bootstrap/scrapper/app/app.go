package app

import (
	"context"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	repository "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/scrapper"
)

type App struct {
	cfg  *config.Config
	repo *repository.Repository
}

func New(cfg *config.Config) *App {
	return &App{
		cfg:  cfg,
		repo: repository.New(),
	}
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	slog.Info("Starting application")
	slog.Debug("Debug level enabled")

	for _, service := range a.services() {
		wg.Add(1)

		go service(ctx, stop, &wg)
	}

	stoppedChan := make(chan struct{})

	go func() {
		wg.Wait()

		stoppedChan <- struct{}{}
	}()

	return a.closer(ctx, a.cfg, stoppedChan)
}
