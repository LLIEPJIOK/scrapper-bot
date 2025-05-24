package app

import (
	"context"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/metrics"
	repo "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/scrapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type InitFunc func(ctx context.Context) error

func (a *App) inits() []InitFunc {
	return []InitFunc{
		a.initDB,
		a.initRepo,
		a.initRedis,
		a.initPrometheus,
	}
}

func (a *App) initDB(ctx context.Context) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		a.cfg.Scrapper.Database.Host,
		a.cfg.Scrapper.Database.Port,
		a.cfg.Scrapper.Database.User,
		a.cfg.Scrapper.Database.Password,
		a.cfg.Scrapper.Database.Name,
		a.cfg.Scrapper.Database.SSLMode,
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
	var err error

	a.repo, err = repo.New(a.db, a.cfg.Scrapper.Database.Type)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	return nil
}

func (a *App) initRedis(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     a.cfg.Scrapper.Redis.Address,
		Password: a.cfg.Scrapper.Redis.Password,
		DB:       a.cfg.Scrapper.Redis.DB,

		DialTimeout:  a.cfg.Scrapper.Redis.DialTimeout,
		ReadTimeout:  a.cfg.Scrapper.Redis.ReadTimeout,
		WriteTimeout: a.cfg.Scrapper.Redis.WriteTimeout,

		PoolSize:     a.cfg.Scrapper.Redis.PoolSize,
		MinIdleConns: a.cfg.Scrapper.Redis.MinIdleConns,
		PoolTimeout:  a.cfg.Scrapper.Redis.PoolTimeout,

		MaxRetries:      a.cfg.Scrapper.Redis.MaxRetries,
		MinRetryBackoff: a.cfg.Scrapper.Redis.MinRetryBackoff,
		MaxRetryBackoff: a.cfg.Scrapper.Redis.MaxRetryBackoff,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	a.rdb = rdb

	return nil
}

func (a *App) initPrometheus(_ context.Context) error {
	a.Prometheus = metrics.NewPrometheus("scrapper")

	return nil
}
