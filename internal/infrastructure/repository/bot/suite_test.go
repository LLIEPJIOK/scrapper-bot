package bot_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	_ "github.com/lib/pq"
)

const (
	migrationDir = "../../../../migrations/bot"
)

type BotSuite struct {
	suite.Suite
	pool   *pgxpool.Pool
	dbcont *testpostgres.PostgresContainer
	db     *sql.DB
}

func (s *BotSuite) BeforeAll(t provider.T) {
	ctx := context.Background()

	var (
		dbName = "postgres"
		dbUser = "postgres"
		dbPass = "postgres"
	)

	postgresContainer, err := testpostgres.Run(ctx,
		"postgres:latest",
		testpostgres.WithDatabase(dbName),
		testpostgres.WithUsername(dbUser),
		testpostgres.WithPassword(dbPass),
		testpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err, "failed to start postgres container")

	var db *sql.DB

	dsn, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err, "failed to get connection string")

	dsn += "sslmode=disable"

	db, err = sql.Open("postgres", dsn)
	require.NoError(t, err, "failed to open db")

	err = db.Ping()
	require.NoError(t, err, "failed to ping db")

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "failed to connect to db")

	err = pool.Ping(ctx)
	require.NoError(t, err, "failed to ping db")

	s.pool = pool
	s.dbcont = postgresContainer
	s.db = db
}

func (s *BotSuite) BeforeEach(t provider.T) {
	ctx := context.Background()

	err := goose.RunContext(ctx, "up", s.db, migrationDir)
	require.NoError(t, err, "failed to run migrations")
}

func (s *BotSuite) AfterEach(t provider.T) {
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx, `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
	`)
	require.NoError(t, err, "failed to recreate schema")
}

func (s *BotSuite) AfterAll(t provider.T) {
	err := s.db.Close()
	require.NoError(t, err, "failed to close db")

	s.pool.Close()

	err = s.dbcont.Terminate(context.Background())
	require.NoError(t, err, "failed to terminate database container")
}

func TestSuite(t *testing.T) {
	suite.RunSuite(t, new(BotSuite))
}
