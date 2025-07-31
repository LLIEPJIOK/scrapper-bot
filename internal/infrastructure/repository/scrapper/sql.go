package scrapper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
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

func (s *SQL) RegisterChat(ctx context.Context, chatID int64) error {
	query := `
		INSERT INTO chats (id) VALUES ($1)
		ON CONFLICT (id) DO UPDATE SET deleted_at = NULL
	`

	_, err := s.db.Exec(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("failed to register chat: %w", err)
	}

	return nil
}

func (s *SQL) DeleteChat(ctx context.Context, chatID int64) error {
	query := `
		UPDATE chats
		SET deleted_at = NOW()
		WHERE id = $1
	`

	cmd, err := s.db.Exec(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return NewErrUnregister(chatID)
	}

	return nil
}

func (s *SQL) TrackLink(ctx context.Context, link *domain.Link) (*domain.Link, error) {
	queryLink := `
		INSERT INTO links (url)
		VALUES ($1)
		ON CONFLICT (url) DO UPDATE SET url = links.url
		RETURNING id
	`

	var id int64

	err := s.db.QueryRow(ctx, queryLink, link.URL).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to add link: %w", err)
	}

	link.ID = id

	queryLinkChat := `
		INSERT INTO links_chats (link_id, chat_id, send_immediately)
		VALUES ($1, $2, $3)
	`

	_, err = s.db.Exec(ctx, queryLinkChat, link.ID, link.ChatID, link.SendImmediately.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to add link chat: %w", err)
	}

	if err := s.addTags(ctx, link); err != nil {
		return nil, err
	}

	if err := s.addFilters(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *SQL) UntrackLink(ctx context.Context, chatID int64, url string) (*domain.Link, error) {
	link, err := s.GetLink(ctx, chatID, url)
	if err != nil {
		return nil, err
	}

	queryFilters := `
		DELETE FROM links_filters
		WHERE link_id = $1 AND chat_id = $2
	`

	_, err = s.db.Exec(ctx, queryFilters, link.ID, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete link filters: %w", err)
	}

	queryTags := `
		DELETE FROM links_tags
		WHERE link_id = $1 AND chat_id = $2
	`

	_, err = s.db.Exec(ctx, queryTags, link.ID, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete link tags: %w", err)
	}

	queryLink := `
		DELETE FROM links_chats
		WHERE link_id = $1 AND chat_id = $2
	`

	_, err = s.db.Exec(ctx, queryLink, link.ID, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete link chat: %w", err)
	}

	return link, nil
}

func (s *SQL) GetLink(ctx context.Context, chatID int64, url string) (*domain.Link, error) {
	link := &domain.Link{}

	queryLink := `
		SELECT id, url
		FROM links
		WHERE url = $1
	`

	err := s.db.QueryRow(ctx, queryLink, url).Scan(&link.ID, &link.URL)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, NewErrLinkNotFound(url)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	tags, err := s.getTags(ctx, link.ID, chatID)
	if err != nil {
		return nil, err
	}

	filters, err := s.getFilters(ctx, link.ID, chatID)
	if err != nil {
		return nil, err
	}

	sendImmediately, err := s.getSendImmediately(ctx, link.ID, chatID)
	if err != nil {
		return nil, err
	}

	link.Tags = tags
	link.Filters = filters
	link.SendImmediately = domain.NewNull(sendImmediately)

	return link, nil
}

func (s *SQL) ListLinks(ctx context.Context, chatID int64) ([]*domain.Link, error) {
	links := []*domain.Link{}

	queryLinks := `
		SELECT l.id, l.url
		FROM links l
		JOIN links_chats lc ON l.id = lc.link_id AND lc.chat_id = $1
	`

	rows, err := s.db.Query(ctx, queryLinks, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	if err := pgxscan.ScanAll(&links, rows); err != nil {
		return nil, fmt.Errorf("failed to scan links: %w", err)
	}

	for i, link := range links {
		tags, err := s.getTags(ctx, link.ID, chatID)
		if err != nil {
			return nil, err
		}

		filters, err := s.getFilters(ctx, link.ID, chatID)
		if err != nil {
			return nil, err
		}

		sendImmediately, err := s.getSendImmediately(ctx, link.ID, chatID)
		if err != nil {
			return nil, err
		}

		links[i].Tags = tags
		links[i].Filters = filters
		links[i].SendImmediately = domain.NewNull(sendImmediately)
	}

	return links, nil
}

func (s *SQL) ListLinksByTag(
	ctx context.Context,
	chatID int64,
	tag string,
) ([]*domain.Link, error) {
	links := []*domain.Link{}

	queryLinks := `
		SELECT l.id, l.url
		FROM links l
		JOIN links_chats lc ON l.id = lc.link_id AND lc.chat_id = $1
		JOIN links_tags lt ON l.id = lt.link_id AND lt.chat_id = $1
		JOIN tags t ON lt.tag_id = t.id AND t.name = $2
	`

	rows, err := s.db.Query(ctx, queryLinks, chatID, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	if err := pgxscan.ScanAll(&links, rows); err != nil {
		return nil, fmt.Errorf("failed to scan links: %w", err)
	}

	for i, link := range links {
		tags, err := s.getTags(ctx, link.ID, chatID)
		if err != nil {
			return nil, err
		}

		filters, err := s.getFilters(ctx, link.ID, chatID)
		if err != nil {
			return nil, err
		}

		sendImmediately, err := s.getSendImmediately(ctx, link.ID, chatID)
		if err != nil {
			return nil, err
		}

		links[i].Tags = tags
		links[i].Filters = filters
		links[i].SendImmediately = domain.NewNull(sendImmediately)
	}

	return links, nil
}

func (s *SQL) GetCheckLinks(
	ctx context.Context,
	from, to time.Time,
	limit uint,
) ([]*domain.CheckLink, error) {
	links := []*domain.CheckLink{}

	queryLinks := `
		SELECT id, url, checked_at
		FROM links l
		WHERE checked_at > $1 AND checked_at <= $2
		ORDER BY checked_at ASC
		LIMIT $3
	`

	rows, err := s.db.Query(ctx, queryLinks, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	if err := pgxscan.ScanAll(&links, rows); err != nil {
		return nil, fmt.Errorf("failed to scan links: %w", err)
	}

	for i, link := range links {
		chats, err := s.getChats(ctx, link.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chats: %w", err)
		}

		links[i].Chats = chats
	}

	return links, nil
}

func (s *SQL) UpdateCheckTime(ctx context.Context, url string, checkedAt time.Time) error {
	queryLink := `
		UPDATE links
		SET checked_at = $1
		WHERE url = $2
	`

	_, err := s.db.Exec(ctx, queryLink, checkedAt, url)
	if err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	return nil
}

func (s *SQL) addTags(ctx context.Context, link *domain.Link) error {
	for _, tag := range link.Tags {
		var tagID int64

		queryTag := `
			INSERT INTO tags (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = tags.name
			RETURNING id
		`

		err := s.db.QueryRow(ctx, queryTag, tag).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("failed to add tag: %w", err)
		}

		queryLinkTag := `
			INSERT INTO links_tags (link_id, tag_id, chat_id)
			VALUES ($1, $2, $3)
		`

		_, err = s.db.Exec(ctx, queryLinkTag, link.ID, tagID, link.ChatID)
		if err != nil {
			return fmt.Errorf("failed to add link tag: %w", err)
		}
	}

	return nil
}

func (s *SQL) getTags(ctx context.Context, linkID, chatID int64) ([]string, error) {
	tags := []string{}

	queryTags := `
		SELECT t.name
		FROM tags t
		JOIN links_tags lt ON t.id = lt.tag_id AND lt.link_id = $1 AND lt.chat_id = $2
	`

	rows, err := s.db.Query(ctx, queryTags, linkID, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	if err := pgxscan.ScanAll(&tags, rows); err != nil {
		return nil, fmt.Errorf("failed to scan tags: %w", err)
	}

	return tags, nil
}

func (s *SQL) addFilters(ctx context.Context, link *domain.Link) error {
	for _, filter := range link.Filters {
		var filterID int64

		queryTag := `
			INSERT INTO filters (value)
			VALUES ($1)
			ON CONFLICT (value) DO UPDATE SET value = filters.value
			RETURNING id
		`

		err := s.db.QueryRow(ctx, queryTag, filter).Scan(&filterID)
		if err != nil {
			return fmt.Errorf("failed to add tag: %w", err)
		}

		queryLinkTag := `
			INSERT INTO links_filters (link_id, filter_id, chat_id)
			VALUES ($1, $2, $3)
		`

		_, err = s.db.Exec(ctx, queryLinkTag, link.ID, filterID, link.ChatID)
		if err != nil {
			return fmt.Errorf("failed to add link filter: %w", err)
		}
	}

	return nil
}

func (s *SQL) getFilters(ctx context.Context, linkID, chatID int64) ([]string, error) {
	filters := []string{}

	queryFilters := `
		SELECT f.value
		FROM filters f
		JOIN links_filters lf ON f.id = lf.filter_id AND lf.link_id = $1 AND lf.chat_id = $2
	`

	rows, err := s.db.Query(ctx, queryFilters, linkID, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get filters: %w", err)
	}

	if err := pgxscan.ScanAll(&filters, rows); err != nil {
		return nil, fmt.Errorf("failed to scan filters: %w", err)
	}

	return filters, nil
}

func (s *SQL) getSendImmediately(ctx context.Context, linkID, chatID int64) (bool, error) {
	var sendImmediately bool

	querySendImmediately := `
		SELECT send_immediately
		FROM links_chats
		WHERE link_id = $1 AND chat_id = $2
	`

	err := s.db.QueryRow(ctx, querySendImmediately, linkID, chatID).Scan(&sendImmediately)
	if err != nil {
		return false, fmt.Errorf("failed to get send immediately: %w", err)
	}

	return sendImmediately, nil
}

func (s *SQL) getChats(ctx context.Context, linkID int64) ([]domain.LinkChat, error) {
	chats := []domain.LinkChat{}

	queryChats := `
		SELECT chat_id, send_immediately
		FROM links_chats
		WHERE link_id = $1
	`

	rows, err := s.db.Query(ctx, queryChats, linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}

	if err := pgxscan.ScanAll(&chats, rows); err != nil {
		return nil, fmt.Errorf("failed to scan chats: %w", err)
	}

	for i, chat := range chats {
		tags, err := s.getTags(ctx, linkID, chat.ChatID)
		if err != nil {
			return nil, err
		}

		filters, err := s.getFilters(ctx, linkID, chat.ChatID)
		if err != nil {
			return nil, err
		}

		chats[i].Tags = tags
		chats[i].Filters = filters
	}

	return chats, nil
}

func (s *SQL) GetActiveLinks(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT
			regexp_replace(url, '^https?://([^/]+)/.*$', '\1') AS host,
			COUNT(*) AS link_count
		FROM links
		GROUP BY host;
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active links: %w", err)
	}

	defer rows.Close()

	activeLinks := make(map[string]int)

	for rows.Next() {
		var (
			host  string
			count int
		)

		if err := rows.Scan(&host, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		activeLinks[host] = count
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows iteration error: %w", rows.Err())
	}

	return activeLinks, nil
}
