package store_test

import (
	"context"
	"go-sqs/apiserver"
	"go-sqs/fixtures"
	"go-sqs/store"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDb(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	ctx := context.Background()
	refreshTokenStore := store.NewRefreshTokenStore(env.DB)
	userStore := store.NewUserStore(env.DB)
	user, err := userStore.CreateUser(ctx, "test@email.com", "password")
	require.NoError(t, err)

	jwtManager := apiserver.NewJwtManager(env.Config)
	userId := uuid.New()
	tokenPair, err := jwtManager.GenerateTokenPair(userId)
	require.NoError(t, err)

	refreshTokenRecord, err := refreshTokenStore.Create(ctx, userId, tokenPair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, user.Id, refreshTokenRecord.UserId)
	expectedExpiration, err := tokenPair.RefreshToken.Claims.GetExpirationTime()
	require.NoError(t, err)
	require.Equal(t, expectedExpiration.Time.UnixMilli(), refreshTokenRecord.ExpiresAt)

	refreshTokenRecord2, err := refreshTokenStore.ByPrimaryKey(ctx, userId, tokenPair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, refreshTokenRecord.UserId, refreshTokenRecord2.UserId)
	require.Equal(t, refreshTokenRecord.HashedToken, refreshTokenRecord2.HashedToken)
	require.Equal(t, refreshTokenRecord.CreatedAt, refreshTokenRecord2.CreatedAt)
	require.Equal(t, refreshTokenRecord.ExpiresAt, refreshTokenRecord2.ExpiresAt)

}
