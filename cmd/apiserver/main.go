package main

import (
	"context"
	"go-sqs/apiserver"
	"go-sqs/config"
	"go-sqs/store"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	os.Setenv("ENV", string(config.Env_Test))
	err := godotenv.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conf, err := config.New()

	if err != nil {
		return err
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)
	db, err := store.NewPostgresDB(conf)
	if err != nil {
		return err
	}
	dataStore := store.New(db)
	jwtManager := apiserver.NewJwtManager(conf)
	server := apiserver.New(conf, logger, dataStore, jwtManager)
	if err = server.Start(ctx); err != nil {
		return err
	}

	return nil
}
