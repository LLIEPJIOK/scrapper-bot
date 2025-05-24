package app

import (
	"context"
	"fmt"

	cache "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/cache/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type InitFunc func(ctx context.Context) error

func (a *App) inits() []InitFunc {
	return []InitFunc{
		a.initDB,
		a.initRedis,
		a.initCache,
		a.initPrometheus,
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

func (a *App) initRedis(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     a.cfg.Bot.Redis.Address,
		Password: a.cfg.Bot.Redis.Password,
		DB:       a.cfg.Bot.Redis.DB,

		DialTimeout:  a.cfg.Bot.Redis.DialTimeout,
		ReadTimeout:  a.cfg.Bot.Redis.ReadTimeout,
		WriteTimeout: a.cfg.Bot.Redis.WriteTimeout,

		PoolSize:     a.cfg.Bot.Redis.PoolSize,
		MinIdleConns: a.cfg.Bot.Redis.MinIdleConns,
		PoolTimeout:  a.cfg.Bot.Redis.PoolTimeout,

		MaxRetries:      a.cfg.Bot.Redis.MaxRetries,
		MinRetryBackoff: a.cfg.Bot.Redis.MinRetryBackoff,
		MaxRetryBackoff: a.cfg.Bot.Redis.MaxRetryBackoff,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	a.rdb = rdb

	return nil
}

func (a *App) initCache(_ context.Context) error {
	a.cache = cache.New(a.rdb, a.cfg.Bot.Redis.DefaultTTL)

	return nil
}

func (a *App) initPrometheus(ctx context.Context) error {
	a.Prometheus = metrics.NewPrometheus("bot")

	return nil
}
