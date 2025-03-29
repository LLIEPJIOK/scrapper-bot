package bot

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Builder struct {
	db *pgxpool.Pool
}

func NewBuilder(db *pgxpool.Pool) *Builder {
	return &Builder{
		db: db,
	}
}

func (s *Builder) AddUpdate(ctx context.Context, update *domain.Update) error {
	query, args, err := sq.Insert("updates").
		Columns("chat_id", "url", "message", "tags").
		Values(update.ChatID, update.URL, update.Message, update.Tags).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build add update query: %w", err)
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to add update: %w", err)
	}

	return nil
}

func (s *Builder) GetUpdatesChats(ctx context.Context, from, to time.Time) ([]int64, error) {
	var chats []int64

	query, args, err := sq.Select("DISTINCT chat_id").
		From("updates").
		Where("created_at > ?", from).
		Where("created_at <= ?", to).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build get updates chats query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates chats: %w", err)
	}

	if err := pgxscan.ScanAll(&chats, rows); err != nil {
		return nil, fmt.Errorf("failed to scan updates chats: %w", err)
	}

	return chats, nil
}

func (s *Builder) GetUpdates(
	ctx context.Context,
	chatID int64,
	from, to time.Time,
) ([]domain.Update, error) {
	var updates []domain.Update

	query, args, err := sq.Select("id", "chat_id", "url", "message", "tags", "created_at").
		From("updates").
		Where(sq.Eq{"chat_id": chatID}).
		Where("created_at > ?", from).
		Where("created_at <= ?", to).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build get updates query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}

	if err := pgxscan.ScanAll(&updates, rows); err != nil {
		return nil, fmt.Errorf("failed to scan updates: %w", err)
	}

	return updates, nil
}
