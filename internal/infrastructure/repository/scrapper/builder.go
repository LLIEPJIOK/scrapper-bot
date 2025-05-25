package scrapper

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
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

func (s *Builder) RegisterChat(ctx context.Context, chatID int64) error {
	query, args, err := sq.Insert("chats").
		Columns("id").
		Values(chatID).
		Suffix("ON CONFLICT (id) DO UPDATE SET deleted_at = NULL").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build register chat query: %w", err)
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to register chat: %w", err)
	}

	return nil
}

func (s *Builder) DeleteChat(ctx context.Context, chatID int64) error {
	query, args, err := sq.Update("chats").
		Set("deleted_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": chatID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete chat query: %w", err)
	}

	cmd, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return NewErrUnregister(chatID)
	}

	return nil
}

func (s *Builder) TrackLink(ctx context.Context, link *domain.Link) (*domain.Link, error) {
	queryLink, args, err := sq.Insert("links").
		Columns("url").
		Values(link.URL).
		Suffix("ON CONFLICT (url) DO UPDATE SET url = links.url RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build link query: %w", err)
	}

	var id int64

	err = s.db.QueryRow(ctx, queryLink, args...).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to add link: %w", err)
	}

	link.ID = id

	queryLinkChat, args, err := sq.Insert("links_chats").
		Columns("link_id", "chat_id", "send_immediately").
		Values(link.ID, link.ChatID, link.SendImmediately.Value).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build link chat query: %w", err)
	}

	_, err = s.db.Exec(ctx, queryLinkChat, args...)
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

func (s *Builder) UntrackLink(ctx context.Context, chatID int64, url string) (*domain.Link, error) {
	link, err := s.GetLink(ctx, chatID, url)
	if err != nil {
		return nil, err
	}

	queries := []struct {
		table string
	}{
		{table: "links_filters"},
		{table: "links_tags"},
		{table: "links_chats"},
	}

	for _, q := range queries {
		query, args, err := sq.Delete(q.table).
			Where(sq.Eq{"link_id": link.ID, "chat_id": chatID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build delete query for %s: %w", q.table, err)
		}

		_, err = s.db.Exec(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to delete from %s: %w", q.table, err)
		}
	}

	return link, nil
}

func (s *Builder) GetLink(ctx context.Context, chatID int64, url string) (*domain.Link, error) {
	link := &domain.Link{}

	query, args, err := sq.Select("id", "url").
		From("links").
		Where(sq.Eq{"url": url}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build get link query: %w", err)
	}

	err = s.db.QueryRow(ctx, query, args...).Scan(&link.ID, &link.URL)
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

func (s *Builder) ListLinks(ctx context.Context, chatID int64) ([]*domain.Link, error) {
	var links []*domain.Link

	query, args, err := sq.Select("l.id", "l.url").
		From("links l").
		Join("links_chats lc ON l.id = lc.link_id AND lc.chat_id = ?", chatID).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list links query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
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

func (s *Builder) ListLinksByTag(
	ctx context.Context,
	chatID int64,
	tag string,
) ([]*domain.Link, error) {
	var links []*domain.Link

	query, args, err := sq.Select("l.id", "l.url").
		From("links l").
		Join("links_chats lc ON l.id = lc.link_id AND lc.chat_id = ?", chatID).
		Join("links_tags lt ON l.id = lt.link_id AND lt.chat_id = ?", chatID).
		Join("tags t ON lt.tag_id = t.id AND t.name = ?", tag).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list links query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
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

func (s *Builder) GetCheckLinks(
	ctx context.Context,
	from, to time.Time,
	limit uint,
) ([]*domain.CheckLink, error) {
	var links []*domain.CheckLink

	query, args, err := sq.Select("id", "url", "checked_at").
		From("links").
		Where("checked_at > ?", from).
		Where("checked_at <= ?", to).
		OrderBy("checked_at ASC").
		Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build get check links query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
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

func (s *Builder) UpdateCheckTime(ctx context.Context, url string, checkedAt time.Time) error {
	query, args, err := sq.Update("links").
		Set("checked_at", checkedAt).
		Where(sq.Eq{"url": url}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update check time query: %w", err)
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	return nil
}

//nolint:dupl // not a duplication
func (s *Builder) addTags(ctx context.Context, link *domain.Link) error {
	for _, tag := range link.Tags {
		var tagID int64

		queryTag, args, err := sq.Insert("tags").
			Columns("name").
			Values(tag).
			Suffix("ON CONFLICT (name) DO UPDATE SET name = tags.name RETURNING id").
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query for tag: %w", err)
		}

		err = s.db.QueryRow(ctx, queryTag, args...).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("failed to add tag: %w", err)
		}

		queryLinkTag, args, err := sq.Insert("links_tags").
			Columns("link_id", "tag_id", "chat_id").
			Values(link.ID, tagID, link.ChatID).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query for link tag: %w", err)
		}

		_, err = s.db.Exec(ctx, queryLinkTag, args...)
		if err != nil {
			return fmt.Errorf("failed to add link tag: %w", err)
		}
	}

	return nil
}

func (s *Builder) getTags(ctx context.Context, linkID, chatID int64) ([]string, error) {
	var tags []string

	queryTags, args, err := sq.Select("t.name").
		From("tags t").
		Join("links_tags lt ON t.id = lt.tag_id AND lt.link_id = ? AND lt.chat_id = ?", linkID, chatID).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query for tags: %w", err)
	}

	rows, err := s.db.Query(ctx, queryTags, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	if err := pgxscan.ScanAll(&tags, rows); err != nil {
		return nil, fmt.Errorf("failed to scan tags: %w", err)
	}

	return tags, nil
}

//nolint:dupl // not a duplication
func (s *Builder) addFilters(ctx context.Context, link *domain.Link) error {
	for _, filter := range link.Filters {
		var filterID int64

		queryFilter, args, err := sq.Insert("filters").
			Columns("value").
			Values(filter).
			Suffix("ON CONFLICT (value) DO UPDATE SET value = filters.value RETURNING id").
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query for filter: %w", err)
		}

		err = s.db.QueryRow(ctx, queryFilter, args...).Scan(&filterID)
		if err != nil {
			return fmt.Errorf("failed to add filter: %w", err)
		}

		queryLinkFilter, args, err := sq.Insert("links_filters").
			Columns("link_id", "filter_id", "chat_id").
			Values(link.ID, filterID, link.ChatID).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query for link filter: %w", err)
		}

		_, err = s.db.Exec(ctx, queryLinkFilter, args...)
		if err != nil {
			return fmt.Errorf("failed to add link filter: %w", err)
		}
	}

	return nil
}

func (s *Builder) getFilters(ctx context.Context, linkID, chatID int64) ([]string, error) {
	var filters []string

	queryFilters, args, err := sq.Select("f.value").
		From("filters f").
		Join("links_filters lf ON f.id = lf.filter_id AND lf.link_id = ? AND lf.chat_id = ?", linkID, chatID).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query for filters: %w", err)
	}

	rows, err := s.db.Query(ctx, queryFilters, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get filters: %w", err)
	}

	if err := pgxscan.ScanAll(&filters, rows); err != nil {
		return nil, fmt.Errorf("failed to scan filters: %w", err)
	}

	return filters, nil
}

func (s *Builder) getSendImmediately(ctx context.Context, linkID, chatID int64) (bool, error) {
	var sendImmediately bool

	querySendImmediately, args, err := sq.Select("send_immediately").
		From("links_chats").
		Where(sq.Eq{"link_id": linkID, "chat_id": chatID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query for send immediately: %w", err)
	}

	err = s.db.QueryRow(ctx, querySendImmediately, args...).Scan(&sendImmediately)
	if err != nil {
		return false, fmt.Errorf("failed to get send immediately: %w", err)
	}

	return sendImmediately, nil
}

func (s *Builder) getChats(ctx context.Context, linkID int64) ([]domain.LinkChat, error) {
	var chats []domain.LinkChat

	queryChats, args, err := sq.Select("chat_id", "send_immediately").
		From("links_chats").
		Where(sq.Eq{"link_id": linkID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query for chats: %w", err)
	}

	rows, err := s.db.Query(ctx, queryChats, args...)
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

func (s *Builder) GetActiveLinks(ctx context.Context) (map[string]int, error) {
	queryBuilder := sq.
		Select(
			`regexp_replace(url, '^https?://([^/]+)/.*$', '\1') AS host`,
			"COUNT(*) AS link_count",
		).
		From("links").
		GroupBy("host")

	sqlStr, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query DB: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return activeLinks, nil
}
