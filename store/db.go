package store

import (
	"context"
	"database/sql"
	"fmt"
	"go-sqs/config"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresDB(conf *config.Config) (*sql.DB, error) {
	dsn := conf.DatabaseUrl()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("faled to open database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
