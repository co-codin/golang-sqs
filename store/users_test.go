package store

import (
	"go-sqs/config"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	os.Setenv("ENV", string(config.Env_Test))

	conf, err := config.New()
	require.NoError(t, err)

	db, err := NewPostgresDB(conf)
	require.NoError(t, err)
	defer db.Close()
}