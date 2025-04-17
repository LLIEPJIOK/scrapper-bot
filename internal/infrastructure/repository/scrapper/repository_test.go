package scrapper_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/scrapper"
	"github.com/stretchr/testify/require"
)

func TestNew_SQL(t *testing.T) {
	t.Parallel()

	repo, err := scrapper.New(nil, "sql")
	require.NoError(t, err, "failed to create repo")

	_, ok := repo.(*scrapper.SQL)
	require.True(t, ok, "repo is not SQL")
}

func TestNew_Builder(t *testing.T) {
	t.Parallel()

	repo, err := scrapper.New(nil, "builder")
	require.NoError(t, err, "failed to create repo")

	_, ok := repo.(*scrapper.Builder)
	require.True(t, ok, "repo is not Builder")
}

func TestNew_Unknown(t *testing.T) {
	t.Parallel()

	repo, err := scrapper.New(nil, "unknown")
	require.ErrorAs(t, err, &scrapper.ErrUnknownDBType{}, "should faild")
	require.Nil(t, repo, "repo should be nil")
}
