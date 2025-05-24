package bot_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	botcache "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/cache/bot"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedisContainer(t *testing.T) (address string, cleanup func()) {
	ctx := context.Background()

	cont, err := testredis.Run(ctx, "redis:7.0")
	require.NoError(t, err, "failed to start redis container")

	dsn, err := cont.ConnectionString(ctx)
	require.NoError(t, err, "failed to get connection string")

	addr := strings.TrimPrefix(dsn, "redis://")

	return addr, func() {
		err := cont.Terminate(ctx)
		require.NoError(t, err, "failed to terminate redis container")
	}
}

func TestCache_ListLinks(t *testing.T) {
	addr, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	defer func() {
		err := rdb.Close()
		require.NoError(t, err, "failed to close redis client")
	}()

	cache := botcache.New(rdb, time.Hour)
	chatID := int64(123)
	list := "test list"

	err := cache.SetListLinks(ctx, chatID, "", list)
	require.NoError(t, err, "failed to set list links")

	gotList, err := cache.GetListLinks(ctx, chatID, "")
	require.NoError(t, err, "failed to get list links")
	assert.Equal(t, list, gotList, "list links do not match")

	_, err = cache.GetListLinks(ctx, chatID, "non-existent")
	assert.Error(t, err, "expected error")
	assert.True(t, errors.As(err, &botcache.ErrNoData{}), "expected ErrNoData error")

	err = cache.InvalidateListLinks(ctx, chatID)
	require.NoError(t, err, "failed to invalidate list links")

	_, err = cache.GetListLinks(ctx, chatID, "")
	assert.Error(t, err, "expected error")
	assert.True(t, errors.As(err, &botcache.ErrNoData{}), "expected ErrNoData error")
}

func TestCache_ListLinksByTag(t *testing.T) {
	addr, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	defer func() {
		err := rdb.Close()
		require.NoError(t, err, "failed to close redis client")
	}()

	cache := botcache.New(rdb, time.Hour)

	chats := []struct {
		chatID int64
		tags   map[string]string
	}{
		{
			chatID: 123,
			tags: map[string]string{
				"tag1": "list1",
				"tag2": "list2",
			},
		},
		{
			chatID: 456,
			tags: map[string]string{
				"tag1": "list3",
				"tag3": "list4",
			},
		},
		{
			chatID: 789,
			tags: map[string]string{
				"tag2": "list5",
				"tag3": "list6",
			},
		},
	}

	for _, chat := range chats {
		for tag, list := range chat.tags {
			err := cache.SetListLinks(ctx, chat.chatID, tag, list)
			require.NoError(
				t,
				err,
				"failed to set list links for chat %d with tag %s",
				chat.chatID,
				tag,
			)
		}
	}

	tagTests := []struct {
		tag      string
		expected map[int64]string
	}{
		{
			tag: "tag1",
			expected: map[int64]string{
				123: "list1",
				456: "list3",
			},
		},
		{
			tag: "tag2",
			expected: map[int64]string{
				123: "list2",
				789: "list5",
			},
		},
		{
			tag: "tag3",
			expected: map[int64]string{
				456: "list4",
				789: "list6",
			},
		},
		{
			tag:      "non-existent",
			expected: map[int64]string{},
		},
	}

	for _, tt := range tagTests {
		for chatID, expectedList := range tt.expected {
			list, err := cache.GetListLinks(ctx, chatID, tt.tag)
			require.NoError(
				t,
				err,
				"failed to get list links for chat %d with tag %s",
				chatID,
				tt.tag,
			)
			assert.Equal(
				t,
				expectedList,
				list,
				"unexpected list for chat %d with tag %s",
				chatID,
				tt.tag,
			)
		}

		_, err := cache.GetListLinks(ctx, 999, tt.tag)
		assert.Error(t, err, "expected error for non-existent chat ID with tag %s", tt.tag)
		assert.True(
			t,
			errors.As(err, &botcache.ErrNoData{}),
			"expected ErrNoData error for non-existent chat ID with tag %s",
			tt.tag,
		)
	}

	err := cache.InvalidateListLinks(ctx, 123)
	require.NoError(t, err, "failed to invalidate list links for chat 123")

	for _, tt := range tagTests {
		_, err := cache.GetListLinks(ctx, 123, tt.tag)
		assert.Error(t, err, "expected error for chat 123 with tag %s after invalidation", tt.tag)
		assert.True(
			t,
			errors.As(err, &botcache.ErrNoData{}),
			"expected ErrNoData error for chat 123 with tag %s after invalidation",
			tt.tag,
		)
	}

	for _, tt := range tagTests {
		for chatID, expectedList := range tt.expected {
			// Skip chat 123 as it was invalidated
			if chatID == 123 {
				continue
			}

			list, err := cache.GetListLinks(ctx, chatID, tt.tag)
			require.NoError(
				t,
				err,
				"failed to get list links for chat %d with tag %s after invalidation",
				chatID,
				tt.tag,
			)
			assert.Equal(
				t,
				expectedList,
				list,
				"unexpected list for chat %d with tag %s after invalidation",
				chatID,
				tt.tag,
			)
		}
	}
}

func TestCache_Updates(t *testing.T) {
	addr, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	defer func() {
		err := rdb.Close()
		require.NoError(t, err, "failed to close redis client")
	}()

	cache := botcache.New(rdb, time.Hour)
	chatID := int64(123)
	update := &domain.Update{
		ChatID: chatID,
		URL:    "https://example.com",
		Message: `{
			"message": "test message",
			"tags": ["tag1", "tag2"]
		}`,
		Tags:            []string{"tag1", "tag2"},
		SendImmediately: domain.NewNull(false),
	}

	err := cache.AddUpdate(ctx, update)
	require.NoError(t, err, "failed to add update")

	chats, err := cache.GetUpdatesChats(ctx)
	require.NoError(t, err, "failed to get updates chats")
	assert.Contains(t, chats, chatID, "chat ID not found in updates chats")

	updates, err := cache.GetAndClearUpdates(ctx, chatID)
	require.NoError(t, err, "failed to get and clear updates")
	assert.Len(t, updates, 1, "expected 1 update")
	assert.Equal(t, chatID, updates[0].ChatID, "chat ID does not match")
	assert.Equal(t, update.URL, updates[0].URL, "URL does not match")
	assert.Equal(t, update.Message, updates[0].Message, "message does not match")
	assert.Equal(t, update.Tags, updates[0].Tags, "tags do not match")
	assert.Equal(
		t,
		update.SendImmediately,
		updates[0].SendImmediately,
		"send immediately does not match",
	)

	chats, err = cache.GetUpdatesChats(ctx)
	require.NoError(t, err, "failed to get updates chats")
	assert.NotContains(t, chats, chatID, "chat ID found in updates chats")
}

func TestCache_MultipleChatsWithUpdates(t *testing.T) {
	addr, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	defer func() {
		err := rdb.Close()
		require.NoError(t, err, "failed to close redis client")
	}()

	cache := botcache.New(rdb, time.Hour)

	chats := []struct {
		chatID  int64
		updates []*domain.Update
	}{
		{
			chatID: 123,
			updates: []*domain.Update{
				{
					ChatID: 123,
					URL:    "https://example.com/1",
					Message: `{
						"message": "test message 1",
						"tags": ["tag1"]
					}`,
					Tags:            []string{"tag1"},
					SendImmediately: domain.NewNull(true),
				},
				{
					ChatID: 123,
					URL:    "https://example.com/2",
					Message: `{
						"message": "test message 2",
						"tags": ["tag2"]
					}`,
					Tags:            []string{"tag2"},
					SendImmediately: domain.NewNull(false),
				},
			},
		},
		{
			chatID: 456,
			updates: []*domain.Update{
				{
					ChatID: 456,
					URL:    "https://example.com/3",
					Message: `{
						"message": "test message 3",
						"tags": ["tag3"]
					}`,
					Tags:            []string{"tag3"},
					SendImmediately: domain.NewNull(true),
				},
			},
		},
		{
			chatID: 789,
			updates: []*domain.Update{
				{
					ChatID: 789,
					URL:    "https://example.com/4",
					Message: `{
						"message": "test message 4",
						"tags": ["tag4"]
					}`,
					Tags:            []string{"tag4"},
					SendImmediately: domain.NewNull(false),
				},
				{
					ChatID: 789,
					URL:    "https://example.com/5",
					Message: `{
						"message": "test message 5",
						"tags": ["tag5"]
					}`,
					Tags:            []string{"tag5"},
					SendImmediately: domain.NewNull(true),
				},
				{
					ChatID: 789,
					URL:    "https://example.com/6",
					Message: `{
						"message": "test message 6",
						"tags": ["tag6"]
					}`,
					Tags:            []string{"tag6"},
					SendImmediately: domain.NewNull(false),
				},
			},
		},
	}

	for _, chat := range chats {
		for _, update := range chat.updates {
			err := cache.AddUpdate(ctx, update)
			require.NoError(t, err, "failed to add update for chat %d", chat.chatID)
		}
	}

	chatIDs, err := cache.GetUpdatesChats(ctx)
	require.NoError(t, err, "failed to get updates chats")
	assert.Len(t, chatIDs, len(chats), "unexpected number of chats")

	for _, chat := range chats {
		assert.Contains(t, chatIDs, chat.chatID, "chat %d not found in updates", chat.chatID)
	}

	for _, chat := range chats {
		updates, err := cache.GetAndClearUpdates(ctx, chat.chatID)
		require.NoError(t, err, "failed to get and clear updates for chat %d", chat.chatID)
		assert.Len(
			t,
			updates,
			len(chat.updates),
			"unexpected number of updates for chat %d",
			chat.chatID,
		)

		for i, update := range updates {
			expected := chat.updates[i]
			assert.Equal(
				t,
				expected.ChatID,
				update.ChatID,
				"chat ID mismatch for update %d in chat %d",
				i,
				chat.chatID,
			)
			assert.Equal(
				t,
				expected.URL,
				update.URL,
				"URL mismatch for update %d in chat %d",
				i,
				chat.chatID,
			)
			assert.Equal(
				t,
				expected.Message,
				update.Message,
				"message mismatch for update %d in chat %d",
				i,
				chat.chatID,
			)
			assert.Equal(
				t,
				expected.Tags,
				update.Tags,
				"tags mismatch for update %d in chat %d",
				i,
				chat.chatID,
			)
			assert.Equal(
				t,
				expected.SendImmediately,
				update.SendImmediately,
				"send immediately mismatch for update %d in chat %d",
				i,
				chat.chatID,
			)
		}
	}

	chatIDs, err = cache.GetUpdatesChats(ctx)
	require.NoError(t, err, "failed to get updates chats after clearing")
	assert.Empty(t, chatIDs, "expected no chats after clearing all updates")
}
