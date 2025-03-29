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
	RegisterChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
	TrackLink(ctx context.Context, link *domain.Link) (*domain.Link, error)
	UntrackLink(ctx context.Context, chatID int64, url string) (*domain.Link, error)
	ListLinks(ctx context.Context, chatID int64) ([]*domain.Link, error)
	GetCheckLinks(
		ctx context.Context,
		from, to time.Time,
		limit int,
	) ([]*domain.CheckLink, error)
	UpdateCheckTime(ctx context.Context, url string, checkedAt time.Time) error
}

type App struct {
	cfg  *config.Config
	db   *pgxpool.Pool
	repo Repository
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
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
