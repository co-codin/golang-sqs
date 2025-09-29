package store

import (
	"context"
	"fmt"
	"go-sqs/config"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal(err)
	}

	os.Setenv("ENV", string(config.Env_Test))

	conf, err := config.New()
	require.NoError(t, err)

	db, err := NewPostgresDB(conf)
	require.NoError(t, err)
	defer db.Close()

	m, err := migrate.New(
		fmt.Sprintf("file:///%s/migrations", conf.ProjectRoot),
		conf.DatabaseUrl())

	require.NoError(t, err)

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	userStore := NewUserStore(db)

	user, err := userStore.CreateUser(context.Background(), "test@test.com", "testingpassword")
	require.NoError(t, err)

	require.Equal(t, "test@test.com", user.Email)
	require.NoError(t, user.ComparePassword("testingpassword"))
}
