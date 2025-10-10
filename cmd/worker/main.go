package main

import (
	"context"
	"go-sqs/config"
	"go-sqs/reports"
	"go-sqs/store"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conf, err := config.New()
	if err != nil {
		return err
	}

	db, err := store.NewPostgresDB(conf)
	if err != nil {
		return err
	}

	dataStore := store.New(db)

	awsConf, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
			},
		}),
	)
	if err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(awsConf, func(options *s3.Options) {
		options.BaseEndpoint = aws.String("http://localhost:4566")
		options.UsePathStyle = true
	})

	sqsClient := sqs.NewFromConfig(awsConf, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String("http://localhost:4566")
	})

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)
	lozClient := reports.NewLozClient(&http.Client{Timeout: time.Second * 10})
	builder := reports.NewReportBuilder(dataStore.ReportStore, lozClient, s3Client, conf, logger)


	maxConcurrency := 2
	worker := reports.NewWorker(conf, builder, logger, sqsClient, maxConcurrency)

	if err := worker.Start(ctx); err != nil {
		return err
	}

	return nil
}
