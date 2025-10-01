package main

import (
	"context"
	"go-sqs/apiserver"
	"go-sqs/config"
	"log"
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

	server := apiserver.New(conf)
	if err = server.Start(ctx); err != nil {
		return err
	}

	return nil
}
