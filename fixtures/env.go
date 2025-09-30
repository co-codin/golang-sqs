package fixtures

import (
	"database/sql"
	"fmt"
	"go-sqs/config"
	"go-sqs/store"
	"os"
	"strings"
	"testing"
	 _ "github.com/lib/pq"
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	Config *config.Config
	DB     *sql.DB
}

func NewTestEnv(t *testing.T) *TestEnv {
	os.Setenv("ENV", string(config.Env_Test))
	err := godotenv.Load("../.env")
	require.NoError(t, err)

	conf, err := config.New()
	require.NoError(t, err)

	db, err := store.NewPostgresDB(conf)
	require.NoError(t, err)

	return &TestEnv{
		Config: conf,
		DB:     db,
	}
}

func (te *TestEnv) SetupDb(t *testing.T) func(t *testing.T) {
	m, err := migrate.New(
		fmt.Sprintf("file:///%s/migrations", te.Config.ProjectRoot),
		te.Config.DatabaseUrl())

	require.NoError(t, err)

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	return te.TeardownDb
}

func (te *TestEnv) TeardownDb(t *testing.T) {
	_, err := te.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", strings.Join([]string{"users", "refresh_tokens", "reports"}, ", ")))
	require.NoError(t, err)
}
