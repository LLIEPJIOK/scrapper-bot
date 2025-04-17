package scrapper_test

import (
	"context"
	"sort"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/scrapper"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ScrapperSuite) TestRegisterAndDeleteChat_Builder(t provider.T) {
	ctx := context.Background()

	repo := scrapper.NewBuilder(s.pool)

	chatID := int64(42)

	err := repo.DeleteChat(ctx, chatID)
	require.ErrorAs(t, err, &scrapper.ErrUnregister{}, "chat should not be registered")

	err = repo.RegisterChat(ctx, chatID)
	require.NoError(t, err, "failed to register chat")

	reg, err := isChatRegistered(ctx, s.pool, chatID)
	require.NoError(t, err, "failed to check if chat is registered")
	assert.True(t, reg, "chat should be registered")

	err = repo.DeleteChat(ctx, chatID)
	require.NoError(t, err, "failed to delete chat")

	reg, err = isChatRegistered(ctx, s.pool, chatID)
	require.NoError(t, err, "failed to check if chat is registered")
	assert.False(t, reg, "chat should be deleted")

	err = repo.RegisterChat(ctx, chatID)
	require.NoError(t, err, "failed to register chat")

	reg, err = isChatRegistered(ctx, s.pool, chatID)
	require.NoError(t, err, "failed to check if chat is registered")
	assert.True(t, reg, "chat should be registered")
}

func (s *ScrapperSuite) TestTrackUntrackLink_Builder(t provider.T) {
	ctx := context.Background()

	repo := scrapper.NewBuilder(s.pool)
	chatID := int64(1)

	err := repo.RegisterChat(ctx, chatID)
	require.NoError(t, err, "failed to register chat")

	link := &domain.Link{
		URL:     "https://example.com",
		ChatID:  chatID,
		Tags:    []string{"news", "tech"},
		Filters: []string{"lang=en", "category=it"},
	}

	t.Run("track new link", func(t provider.T) {
		tracked, err := repo.TrackLink(ctx, link)
		require.NoError(t, err, "failed to track link")
		assert.NotZero(t, tracked.ID, "link should be tracked")

		link.ID = tracked.ID

		var count int
		err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM links WHERE url = $1", link.URL).
			Scan(&count)
		require.NoError(t, err, "failed to count links")
		assert.Equal(t, 1, count, "link should be tracked")

		err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM links_chats WHERE link_id = $1 AND chat_id = $2", tracked.ID, chatID).
			Scan(&count)
		require.NoError(t, err, "failed to count links_chats")
		assert.Equal(t, 1, count, "link should be tracked for chat")

		var tags []string

		rows, err := s.pool.Query(
			ctx,
			"SELECT t.name FROM tags t JOIN links_tags lt ON t.id = lt.tag_id WHERE lt.link_id = $1",
			tracked.ID,
		)
		require.NoError(t, err, "failed to scan tags")

		err = pgxscan.ScanAll(&tags, rows)
		require.NoError(t, err, "failed to scan tags")

		assert.ElementsMatch(t, link.Tags, tags, "tags should be tracked")

		var filters []string

		rows, err = s.pool.Query(
			ctx,
			"SELECT f.value FROM filters f JOIN links_filters lf ON f.id = lf.filter_id WHERE lf.link_id = $1",
			tracked.ID,
		)
		require.NoError(t, err, "failed to scan filters")

		err = pgxscan.ScanAll(&filters, rows)
		require.NoError(t, err, "failed to scan filters")

		assert.ElementsMatch(t, link.Filters, filters, "filters should be tracked")
	})

	t.Run("untrack link", func(t provider.T) {
		untracked, err := repo.UntrackLink(ctx, chatID, link.URL)
		require.NoError(t, err, "failed to untrack link")
		assert.Equal(t, link.ID, untracked.ID, "link id should be equal")
		assert.Equal(t, link.URL, untracked.URL, "link url should be equal")

		var count int
		err = s.pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM links_chats WHERE link_id = $1 AND chat_id = $2",
			untracked.ID, chatID,
		).Scan(&count)
		require.NoError(t, err, "failed to count links_chats")
		assert.Zero(t, count, "link should be untracked for chat")

		err = s.pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM links_tags WHERE link_id = $1 AND chat_id = $2",
			untracked.ID, chatID,
		).Scan(&count)
		require.NoError(t, err, "failed to count links_tags")
		assert.Zero(t, count, "link should be untracked for chat")
	})
}

func (s *ScrapperSuite) TestGetLink_Builder(t provider.T) {
	ctx := context.Background()

	repo := scrapper.NewBuilder(s.pool)
	chatID := int64(1)

	err := repo.RegisterChat(ctx, chatID)
	require.NoError(t, err, "failed to register chat")

	link := &domain.Link{
		URL:     "https://example.org",
		ChatID:  chatID,
		Tags:    []string{"blog"},
		Filters: []string{"author=john"},
	}

	tracked, err := repo.TrackLink(ctx, link)
	require.NoError(t, err, "failed to track link")

	t.Run("get existing link", func(t provider.T) {
		found, err := repo.GetLink(ctx, chatID, link.URL)
		require.NoError(t, err, "failed to get link")
		require.Equal(t, tracked.ID, found.ID, "link id should be equal")
		require.Equal(t, link.URL, found.URL, "link url should be equal")
		require.ElementsMatch(t, link.Tags, found.Tags, "tags should be equal")
		require.ElementsMatch(t, link.Filters, found.Filters, "filters should be equal")
	})

	t.Run("get non-existent link", func(t provider.T) {
		_, err := repo.GetLink(ctx, chatID, "invalid-url")
		require.ErrorAs(t, err, &scrapper.ErrLinkNotFound{}, "link should not be found")
	})
}

func (s *ScrapperSuite) TestListLinks_Builder(t provider.T) {
	ctx := context.Background()

	repo := scrapper.NewBuilder(s.pool)
	chatID := int64(1)

	err := repo.RegisterChat(ctx, chatID)
	require.NoError(t, err, "failed to register chat")

	links := []*domain.Link{
		{URL: "https://link1.com", ChatID: chatID, Tags: []string{"t1"}},
		{URL: "https://link2.com", ChatID: chatID, Filters: []string{"f1"}},
	}

	for _, l := range links {
		_, err := repo.TrackLink(ctx, l)
		require.NoError(t, err)
	}

	t.Run("list links for chat", func(t provider.T) {
		result, err := repo.ListLinks(ctx, chatID)
		require.NoError(t, err)
		require.Len(t, result, 2)

		sort.Slice(result, func(i, j int) bool {
			return result[i].URL < result[j].URL
		})

		assert.Equal(t, links[0].URL, result[0].URL, "link url should be equal")
		assert.ElementsMatch(t, links[0].Tags, result[0].Tags, "link tags should be equal")
		assert.ElementsMatch(t, links[0].Filters, result[0].Filters, "link filters should be equal")

		assert.Equal(t, links[1].URL, result[1].URL, "link url should be equal")
		assert.ElementsMatch(t, links[1].Tags, result[1].Tags, "link tags should be equal")
		assert.ElementsMatch(t, links[1].Filters, result[1].Filters, "link filters should be equal")
	})
}

func (s *ScrapperSuite) TestGetCheckLinks_Builder(t provider.T) {
	ctx := context.Background()

	repo := scrapper.NewBuilder(s.pool)
	chat1ID := int64(1)
	chat2ID := int64(2)

	err := repo.RegisterChat(ctx, chat1ID)
	require.NoError(t, err, "failed to register chat")
	err = repo.RegisterChat(ctx, chat2ID)
	require.NoError(t, err, "failed to register chat")

	links := []*domain.Link{
		{URL: "https://link1.com", ChatID: chat1ID, Tags: []string{"t1"}},
		{URL: "https://link2.com", ChatID: chat1ID, Filters: []string{"f1"}},
		{URL: "https://link1.com", ChatID: chat2ID, Tags: []string{"t1"}},
	}

	for _, l := range links {
		_, err := repo.TrackLink(ctx, l)
		require.NoError(t, err)
	}

	err = repo.UpdateCheckTime(ctx, links[0].URL, time.Now().Add(-48*time.Hour))
	require.NoError(t, err, "failed to update check time")
	err = repo.UpdateCheckTime(ctx, links[1].URL, time.Now().Add(-1*time.Hour))
	require.NoError(t, err, "failed to update check time")

	tm := time.Now().Add(-time.Minute)

	checkLinks, err := repo.GetCheckLinks(ctx, tm.Add(-50*time.Hour), tm, 10)
	require.NoError(t, err, "failed to get links")
	require.Len(t, checkLinks, 2, "should get 2 link")

	sort.Slice(checkLinks, func(i, j int) bool {
		return checkLinks[i].URL < checkLinks[j].URL
	})

	assert.Equal(t, links[0].URL, checkLinks[0].URL, "link url should be equal")
	assert.Len(t, checkLinks[0].Chats, 2, "link should be checked for 2 chats")

	sort.Slice(checkLinks[0].Chats, func(i, j int) bool {
		return checkLinks[0].Chats[i].ChatID < checkLinks[0].Chats[j].ChatID
	})

	assert.Equal(t, chat1ID, checkLinks[0].Chats[0].ChatID, "chat id should be equal")
	assert.ElementsMatch(t, links[0].Tags, checkLinks[0].Chats[0].Tags, "link tags should be equal")
	assert.ElementsMatch(
		t,
		links[0].Filters,
		checkLinks[0].Chats[0].Filters,
		"link filters should be equal",
	)

	assert.Equal(t, chat2ID, checkLinks[0].Chats[1].ChatID, "chat id should be equal")
	assert.ElementsMatch(t, links[0].Tags, checkLinks[0].Chats[1].Tags, "link tags should be equal")
	assert.ElementsMatch(
		t,
		links[0].Filters,
		checkLinks[0].Chats[1].Filters,
		"link filters should be equal",
	)

	assert.Equal(t, links[1].URL, checkLinks[1].URL, "link url should be equal")
	assert.Len(t, checkLinks[1].Chats, 1, "link should be checked for 1 chat")
	assert.Equal(t, chat1ID, checkLinks[1].Chats[0].ChatID, "chat id should be equal")
	assert.ElementsMatch(t, links[1].Tags, checkLinks[1].Chats[0].Tags, "link tags should be equal")
	assert.ElementsMatch(
		t,
		links[1].Filters,
		checkLinks[1].Chats[0].Filters,
		"link filters should be equal",
	)

	err = repo.UpdateCheckTime(ctx, links[0].URL, time.Now())
	require.NoError(t, err, "failed to update check time")

	checkLinks, err = repo.GetCheckLinks(ctx, tm.Add(-24*time.Hour), tm, 10)
	require.NoError(t, err, "failed to get links")
	require.Len(t, checkLinks, 1, "should get 1 link")

	assert.Equal(t, links[1].URL, checkLinks[0].URL, "link url should be equal")
	assert.Len(t, checkLinks[0].Chats, 1, "link should be checked for 1 chat")
	assert.Equal(t, chat1ID, checkLinks[0].Chats[0].ChatID, "chat id should be equal")
	assert.ElementsMatch(t, links[1].Tags, checkLinks[0].Chats[0].Tags, "link tags should be equal")
	assert.ElementsMatch(
		t,
		links[1].Filters,
		checkLinks[0].Chats[0].Filters,
		"link filters should be equal",
	)
}

func (s *ScrapperSuite) TestGetCheckLinks_Pagination_Builder(t provider.T) {
	ctx := context.Background()

	repo := scrapper.NewBuilder(s.pool)
	chat1ID := int64(1)
	chat2ID := int64(2)

	err := repo.RegisterChat(ctx, chat1ID)
	require.NoError(t, err, "failed to register chat")
	err = repo.RegisterChat(ctx, chat2ID)
	require.NoError(t, err, "failed to register chat")

	links := []*domain.Link{
		{URL: "https://link1.com", ChatID: chat1ID, Tags: []string{"t1"}},
		{URL: "https://link2.com", ChatID: chat1ID, Filters: []string{"f1"}},
		{URL: "https://link1.com", ChatID: chat2ID, Tags: []string{"t1"}},
	}

	for _, l := range links {
		_, err := repo.TrackLink(ctx, l)
		require.NoError(t, err)
	}

	err = repo.UpdateCheckTime(ctx, links[0].URL, time.Now().Add(-48*time.Hour))
	require.NoError(t, err, "failed to update check time")
	err = repo.UpdateCheckTime(ctx, links[1].URL, time.Now().Add(-1*time.Hour))
	require.NoError(t, err, "failed to update check time")

	tm := time.Now().Add(-time.Minute)

	checkLinks, err := repo.GetCheckLinks(ctx, tm.Add(-50*time.Hour), tm, 1)
	require.NoError(t, err, "failed to get links")
	require.Len(t, checkLinks, 1, "should get 2 link")

	assert.Equal(t, links[0].URL, checkLinks[0].URL, "link url should be equal")
	assert.Len(t, checkLinks[0].Chats, 2, "link should be checked for 2 chats")

	assert.Equal(t, chat1ID, checkLinks[0].Chats[0].ChatID, "chat id should be equal")
	assert.ElementsMatch(t, links[0].Tags, checkLinks[0].Chats[0].Tags, "link tags should be equal")
	assert.ElementsMatch(
		t,
		links[0].Filters,
		checkLinks[0].Chats[0].Filters,
		"link filters should be equal",
	)

	assert.Equal(t, chat2ID, checkLinks[0].Chats[1].ChatID, "chat id should be equal")
	assert.ElementsMatch(t, links[0].Tags, checkLinks[0].Chats[1].Tags, "link tags should be equal")
	assert.ElementsMatch(
		t,
		links[0].Filters,
		checkLinks[0].Chats[1].Filters,
		"link filters should be equal",
	)
}
