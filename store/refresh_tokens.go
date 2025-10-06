package store

import (
	"context"
	"database/sql"
	"encoding/base64"

	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshToken struct {
	UserId      int64  `db:"user_id"`
	HashedToken string `db:"hashed_token"`
	ExpiresAt   int64  `db:"expires_at"`
	CreatedAt   int64  `db:"created_at"`
}

func (s *RefreshTokenStore) Create(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const insert = `
		INSERT INTO refresh_tokens (user_id, hashed_token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token.Raw), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}
	base64TokenHash := base64.StdEncoding.EncodeToString(hashedToken)

	var refreshToken RefreshToken
	if err := s.db.GetContext(ctx, &refreshToken, insert, userId, base64TokenHash); err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &refreshToken, nil
}

