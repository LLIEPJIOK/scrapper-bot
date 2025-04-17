package bot_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/bot"
	"github.com/stretchr/testify/require"
)

func TestNew_SQL(t *testing.T) {
	t.Parallel()

	repo, err := bot.New(nil, "sql")
	require.NoError(t, err, "failed to create repo")

	_, ok := repo.(*bot.SQL)
	require.True(t, ok, "repo is not SQL")
}

func TestNew_Builder(t *testing.T) {
	t.Parallel()

	repo, err := bot.New(nil, "builder")
	require.NoError(t, err, "failed to create repo")

	_, ok := repo.(*bot.Builder)
	require.True(t, ok, "repo is not Builder")
}

func TestNew_Unknown(t *testing.T) {
	t.Parallel()

	repo, err := bot.New(nil, "unknown")
	require.ErrorAs(t, err, &bot.ErrUnknownDBType{}, "should faild")
	require.Nil(t, repo, "repo should be nil")
}
