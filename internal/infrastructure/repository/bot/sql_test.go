package bot_test

import (
	"context"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/bot"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/require"
)

func (s *BotSuite) TestAddAndGetUpdates_SQL(t provider.T) {
	ctx := context.Background()

	repo := bot.NewSQL(s.pool)

	update1 := &domain.Update{
		ChatID:  1,
		URL:     "http://example.com",
		Message: "Test message",
		Tags:    []string{"tag1", "tag2"},
	}

	err := repo.AddUpdate(ctx, update1)
	require.NoError(t, err)

	now := time.Now()
	from := now.Add(-1 * time.Minute)
	to := now.Add(1 * time.Minute)

	chats, err := repo.GetUpdatesChats(ctx, from, to)
	require.NoError(t, err)
	require.Len(t, chats, 1)
	require.Equal(t, int64(1), chats[0])

	updates, err := repo.GetUpdates(ctx, 1, from, to)
	require.NoError(t, err)
	require.Len(t, updates, 1)
	require.Equal(t, update1.URL, updates[0].URL)
	require.Equal(t, update1.Message, updates[0].Message)
	require.Equal(t, update1.Tags, updates[0].Tags)
}

func (s *BotSuite) TestGetUpdatesEmpty_SQL(t provider.T) {
	ctx := context.Background()

	sqlRepo := bot.NewSQL(s.pool)

	from := time.Now().Add(-10 * time.Minute)
	to := time.Now().Add(-5 * time.Minute)

	chats, err := sqlRepo.GetUpdatesChats(ctx, from, to)
	require.NoError(t, err)
	require.Empty(t, chats)

	updates, err := sqlRepo.GetUpdates(ctx, 1, from, to)
	require.NoError(t, err)
	require.Empty(t, updates)
}

func (s *BotSuite) TestGetUpdatesMultipleChats_SQL(t provider.T) {
	ctx := context.Background()

	sqlRepo := bot.NewSQL(s.pool)

	updatesToAdd := []*domain.Update{
		{ChatID: 1, URL: "http://example.com/1", Message: "Message 1", Tags: []string{"a"}},
		{ChatID: 2, URL: "http://example.com/2", Message: "Message 2", Tags: []string{"b"}},
		{ChatID: 1, URL: "http://example.com/3", Message: "Message 3", Tags: []string{"c"}},
		{ChatID: 3, URL: "http://example.com/4", Message: "Message 4", Tags: []string{"d"}},
	}
	for _, upd := range updatesToAdd {
		err := sqlRepo.AddUpdate(ctx, upd)
		require.NoError(t, err)
	}

	now := time.Now()
	from := now.Add(-1 * time.Minute)
	to := now.Add(1 * time.Minute)

	chats, err := sqlRepo.GetUpdatesChats(ctx, from, to)
	require.NoError(t, err)
	require.ElementsMatch(t, []int64{1, 2, 3}, chats)
}

func (s *BotSuite) TestEdgeCaseTimeFiltering_SQL(t provider.T) {
	ctx := context.Background()

	sqlRepo := bot.NewSQL(s.pool)

	now := time.Now()
	update := &domain.Update{
		ChatID:  1,
		URL:     "http://edgecase.com",
		Message: "Edge case message",
		Tags:    []string{"edge"},
	}
	err := sqlRepo.AddUpdate(ctx, update)
	require.NoError(t, err)

	updates, err := sqlRepo.GetUpdates(ctx, 1, now.Add(-1*time.Minute), now.Add(1*time.Minute))
	require.NoError(t, err)
	require.Len(t, updates, 1)

	createdAt := updates[0].CreatedAt
	from := createdAt.Add(-1 * time.Second)
	to := createdAt

	filtered, err := sqlRepo.GetUpdates(ctx, 1, from, to)
	require.NoError(t, err)
	require.Len(t, filtered, 1)
}
