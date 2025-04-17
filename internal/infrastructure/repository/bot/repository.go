package bot

import (
	"context"
	"time"

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

func New(db *pgxpool.Pool, tpe string) (Repository, error) {
	switch tpe {
	case "sql":
		return NewSQL(db), nil

	case "builder":
		return NewBuilder(db), nil

	default:
		return nil, NewErrUnknownDBType(tpe)
	}
}
