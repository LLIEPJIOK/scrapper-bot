package app

import (
	"context"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddUpdate(ctx context.Context, update *domain.Update) error
	GetUpdatesChats(ctx context.Context, from, to time.Time) ([]int64, error)
	GetUpdates(
		ctx context.Context,
		chatID int64,
		from, to time.Time,
	) ([]domain.Update, error)
}

type App struct {
	cfg      *config.Config
	channels *domain.Channels
	db       *pgxpool.Pool
	repo     Repository
}

func New(cfg *config.Config) *App {
	return &App{
		cfg:      cfg,
		channels: domain.NewChannels(),
	}
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for _, init := range a.inits() {
		if err := init(ctx); err != nil {
			return err
		}
	}

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
