package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"go-sqs/config"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Worker struct {
	config      *config.Config
	builder     *ReportBuilder
	logger      *slog.Logger
	sqsClient   *sqs.Client
	channel     chan types.Message
	concurrency int
}

func NewWorker(cfg *config.Config, builder *ReportBuilder, logger *slog.Logger, sqsClient *sqs.Client, maxConcurrency int) *Worker {
	return &Worker{
		config:      cfg,
		builder:     builder,
		logger:      logger,
		sqsClient:   sqsClient,
		channel:     make(chan types.Message, maxConcurrency),
		concurrency: maxConcurrency,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	queueUrlOutput, err := w.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &w.config.SqsQueue,
	})
	if err != nil {
		return fmt.Errorf("failed to get SQS queue URL: %w", err)
	}
	w.logger.Info("SQS queue URL retrieved", slog.String("queueUrl", *queueUrlOutput.QueueUrl))

	for i := 0; i < w.concurrency; i++ {
		go func(id int) {
			for {
				select {
				case <-ctx.Done():
					w.logger.Error("Worker shutting down", slog.Int("workerId", id))
					return
				case message := <-w.channel:
					if err := w.processMessage(ctx, message, queueUrlOutput.QueueUrl); err != nil {
						w.logger.Error("failed to process message", "error", err)
					}
					w.logger.Info("Worker processing message", slog.Int("workerId", id), slog.String("messageId", *message.MessageId))

				}
			}
		}(i)
	}

	for {
		output, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            queueUrlOutput.QueueUrl,
			MaxNumberOfMessages: int32(w.concurrency),
			WaitTimeSeconds:     10,
		})
		if err != nil {
			w.logger.Error("failed to receive messages", "error", err)
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}

		if len(output.Messages) == 0 {
			continue
		}

		for _, message := range output.Messages {
			w.channel <- message
		}
	}

}

func (w *Worker) processMessage(ctx context.Context, message types.Message, queueUrl *string) error {
	if message.Body == nil || *message.Body == "" {
		w.logger.Warn("Received empty message body", slog.String("messageId", *message.MessageId))
		return nil
	}

	var msg SqsMessage
	if err := json.Unmarshal([]byte(*message.Body), &msg); err != nil {
		w.logger.Error("failed to unmarshal message body", "error", err, slog.String("messageId", *message.MessageId))
		return nil
	}

	builderCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := w.builder.Build(builderCtx, msg.UserId, msg.ReportId)
	if err != nil {
		return fmt.Errorf("failed to build report: %w", err)
	}

	_, err = w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      queueUrl,
		ReceiptHandle: message.ReceiptHandle,
	})

	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	w.logger.Info("Successfully processed and deleted message", slog.String("messageId", *message.MessageId))
	return nil
}
