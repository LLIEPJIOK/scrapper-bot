package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/redis/go-redis/v9"
)

const pageSize = 100

type Cache struct {
	rdb        *redis.Client
	defaultTTL time.Duration
}

func New(rdb *redis.Client, defaultTTL time.Duration) *Cache {
	return &Cache{
		rdb:        rdb,
		defaultTTL: defaultTTL,
	}
}

func (c *Cache) GetListLinks(
	ctx context.Context,
	chatID int64,
	tag string,
) (string, error) {
	key := c.getListLinksKey(chatID, tag)

	cmd := c.rdb.Get(ctx, key)
	if errors.Is(cmd.Err(), redis.Nil) {
		return "", NewErrNoData()
	}

	if cmd.Err() != nil {
		return "", fmt.Errorf("failed to get list links: %w", cmd.Err())
	}

	return cmd.Val(), nil
}

func (c *Cache) SetListLinks(
	ctx context.Context,
	chatID int64,
	tag string,
	list string,
) error {
	key := c.getListLinksKey(chatID, tag)

	if err := c.rdb.Set(ctx, key, list, c.defaultTTL).Err(); err != nil {
		return fmt.Errorf("failed to set list links: %w", err)
	}

	return nil
}

func (c *Cache) InvalidateListLinks(ctx context.Context, chatID int64) error {
	if err := c.rdb.Del(ctx, "links:"+strconv.FormatInt(chatID, 10)).Err(); err != nil {
		return fmt.Errorf("failed to delete links: %w", err)
	}

	// delete all lists with tags
	var cursor uint64

	match := fmt.Sprintf("links:%d:*", chatID)

	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, match, pageSize).Result()
		if err != nil {
			return fmt.Errorf("failed to scan links: %w", err)
		}

		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete links: %w", err)
			}
		}

		cursor = next

		if cursor == 0 {
			break
		}
	}

	return nil
}

func (c *Cache) AddUpdate(ctx context.Context, update *domain.Update) error {
	chatsKey := "updates:chats"
	updatesKey := fmt.Sprintf("updates:chat:%d", update.ChatID)

	raw, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	pipe := c.rdb.Pipeline()
	_ = pipe.SAdd(ctx, chatsKey, update.ChatID)
	_ = pipe.RPush(ctx, updatesKey, raw)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to add update: %w", err)
	}

	return nil
}

func (c *Cache) GetUpdatesChats(ctx context.Context) ([]int64, error) {
	rawIDs, err := c.rdb.SMembers(ctx, "updates:chats").Result()
	if err != nil {
		return nil, err
	}

	chats := make([]int64, 0, len(rawIDs))

	for _, rawID := range rawIDs {
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			slog.Error(
				"failed to parse chat id",
				slog.Any("id", rawID),
				slog.Any("service", "cache"),
				slog.Any("error", err),
			)
		}

		chats = append(chats, id)
	}

	return chats, nil
}

func (c *Cache) GetAndClearUpdates(
	ctx context.Context,
	chatID int64,
) ([]*domain.Update, error) {
	chatsKey := "updates:chats"
	updatesKey := fmt.Sprintf("updates:chat:%d", chatID)

	pipe := c.rdb.TxPipeline()
	lcmd := pipe.LRange(ctx, updatesKey, 0, -1)
	_ = pipe.Del(ctx, updatesKey)
	_ = pipe.SRem(ctx, chatsKey, chatID)

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to get and clear updates: %w", err)
	}

	rawUpdates, err := lcmd.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get and clear updates: %w", err)
	}

	updates := make([]*domain.Update, 0, len(rawUpdates))

	for _, rawUpdate := range rawUpdates {
		var update domain.Update

		if err := json.Unmarshal([]byte(rawUpdate), &update); err != nil {
			return nil, fmt.Errorf("failed to unmarshal update: %w", err)
		}

		updates = append(updates, &update)
	}

	return updates, nil
}

func (c *Cache) getListLinksKey(chatID int64, tag string) string {
	if tag == "" {
		return "links:" + strconv.FormatInt(chatID, 10)
	}

	return fmt.Sprintf("links:%d:%s", chatID, tag)
}
