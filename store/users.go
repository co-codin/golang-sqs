package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type User struct {
	Id                   uuid.UUID `db:"id"`
	Email                string    `db:"email"`
	HashedPasswordBase64 string    `db:"hashed_password"`
	CreatedAt            time.Time `db:"created_at"`
}

func (u *User) ComparePassword(password string) error {
	hashedPassword, err := base64.StdEncoding.DecodeString(u.HashedPasswordBase64)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) CreateUser(ctx context.Context, email, password string) (*User, error) {
	const dml = `INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING *;`
	var user User

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashedPasswordBase64 := base64.StdEncoding.EncodeToString(bytes)

	if err := s.db.GetContext(ctx, &user, dml, email, hashedPasswordBase64); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) ByEmail(ctx context.Context, email string) (*User, error) {
	const query = `SELECT * FROM users where email = $1;`
	var user User
	if err := s.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) ByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	const query = `SELECT * FROM users where id = $1;`
	var user User
	if err := s.db.GetContext(ctx, &user, query, userID); err != nil {
		return nil, err
	}
	return &user, nil
}
