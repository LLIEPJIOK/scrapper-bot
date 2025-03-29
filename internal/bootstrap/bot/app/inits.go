package app

import (
	"context"
	"fmt"

	repo "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/bot"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InitFunc func(ctx context.Context) error

func (a *App) inits() []InitFunc {
	return []InitFunc{
		a.initDB,
		a.initRepo,
	}
}

func (a *App) initDB(ctx context.Context) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		a.cfg.Bot.Database.Host,
		a.cfg.Bot.Database.Port,
		a.cfg.Bot.Database.User,
		a.cfg.Bot.Database.Password,
		a.cfg.Bot.Database.Name,
		a.cfg.Bot.Database.SSLMode,
	)

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create db: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}

	a.db = db

	return nil
}

func (a *App) initRepo(_ context.Context) error {
	switch a.cfg.Scrapper.Database.Type {
	case "sql":
		a.repo = repo.NewSQL(a.db)

	case "builder":
		// TODO: create repository

	default:
		return NewErrUnknownDBType(a.cfg.Scrapper.Database.Type)
	}

	return nil
}
