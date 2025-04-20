package app

import (
	"context"
	"fmt"

	cache "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/cache/bot"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type InitFunc func(ctx context.Context) error

func (a *App) inits() []InitFunc {
	return []InitFunc{
		a.initDB,
		a.initRedis,
		a.initCache,
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
		Addr:     a.cfg.Redis.Address,
		Password: a.cfg.Redis.Password,
		DB:       a.cfg.Redis.DB,

		DialTimeout:  a.cfg.Redis.DialTimeout,
		ReadTimeout:  a.cfg.Redis.ReadTimeout,
		WriteTimeout: a.cfg.Redis.WriteTimeout,

		PoolSize:     a.cfg.Redis.PoolSize,
		MinIdleConns: a.cfg.Redis.MinIdleConns,
		PoolTimeout:  a.cfg.Redis.PoolTimeout,

		MaxRetries:      a.cfg.Redis.MaxRetries,
		MinRetryBackoff: a.cfg.Redis.MinRetryBackoff,
		MaxRetryBackoff: a.cfg.Redis.MaxRetryBackoff,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	a.rdb = rdb

	return nil
}

func (a *App) initCache(_ context.Context) error {
	a.cache = cache.New(a.rdb, a.cfg.Redis.DefaultTTL)

	return nil
}
