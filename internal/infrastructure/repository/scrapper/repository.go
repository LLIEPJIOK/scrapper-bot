package scrapper

import (
	"context"
	"time"

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
		limit uint,
	) ([]*domain.CheckLink, error)
	UpdateCheckTime(ctx context.Context, url string, checkedAt time.Time) error
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
