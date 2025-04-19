package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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

func (c *Cache) getListLinksKey(chatID int64, tag string) string {
	if tag == "" {
		return "links:" + strconv.FormatInt(chatID, 10)
	}

	return fmt.Sprintf("links:%d:%s", chatID, tag)
}
