package store_test

import (
	"context"
	"go-sqs/fixtures"
	"go-sqs/store"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReportStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDb(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	ctx := context.Background()
	reportStore := store.NewReportStore(env.DB)
	userStore := store.NewUserStore(env.DB)
	user, err := userStore.CreateUser(ctx, "test@test.com", "secretpassword")
	require.NoError(t, err)

	now := time.Now()
	report, err := reportStore.Create(ctx, user.Id, "monsters")
	require.NoError(t, err)
	require.Equal(t, user.Id, report.UserID)
	require.Equal(t, "monsters", report.ReportType)
	require.Less(t, now.UnixNano(), report.CreatedAt.UnixNano())
}
