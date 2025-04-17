package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQL struct {
	db *pgxpool.Pool
}

func NewSQL(db *pgxpool.Pool) *SQL {
	return &SQL{
		db: db,
	}
}

func (s *SQL) AddUpdate(ctx context.Context, update *domain.Update) error {
	query := `
		INSERT INTO updates (chat_id, url, message, tags)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.Exec(ctx, query, update.ChatID, update.URL, update.Message, update.Tags)
	if err != nil {
		return fmt.Errorf("failed to add update: %w", err)
	}

	return nil
}

func (s *SQL) GetUpdatesChats(ctx context.Context, from, to time.Time) ([]int64, error) {
	chats := make([]int64, 0)

	query := `
		SELECT DISTINCT chat_id
		FROM updates u
		WHERE created_at > $1 AND created_at <= $2
	`

	rows, err := s.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}

	if err := pgxscan.ScanAll(&chats, rows); err != nil {
		return nil, fmt.Errorf("failed to scan updates: %w", err)
	}

	return chats, nil
}

func (s *SQL) GetUpdates(
	ctx context.Context,
	chatID int64,
	from, to time.Time,
) ([]domain.Update, error) {
	updates := make([]domain.Update, 0)

	query := `
		SELECT id, chat_id, url, message, tags, created_at
		FROM updates
		WHERE chat_id = $1 AND created_at > $2 AND created_at <= $3
	`

	rows, err := s.db.Query(ctx, query, chatID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}

	if err := pgxscan.ScanAll(&updates, rows); err != nil {
		return nil, fmt.Errorf("failed to scan updates: %w", err)
	}

	return updates, nil
}
