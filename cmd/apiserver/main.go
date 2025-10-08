package main

import (
	"context"
	"fmt"
	"go-sqs/apiserver"
	"go-sqs/config"
	"go-sqs/store"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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

	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     conf.AwsAccessKeyID,
				SecretAccessKey: conf.AwsAccessSecretKey,
			},
		}),
	)
	if err != nil {
		fmt.Println("couldn't load default config")
		fmt.Println(err)
	}

	sqsClient := sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String("http://localhost:4566")
	})

	server := apiserver.New(conf, logger, dataStore, jwtManager, sqsClient)
	if err = server.Start(ctx); err != nil {
		return err
	}

	return nil
}
