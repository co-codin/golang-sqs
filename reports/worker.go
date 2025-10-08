package reports

import (
	"go-sqs/config"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Worker struct {
	config    *config.Config
	builder   *ReportBuilder
	logger    *slog.Logger
	sqsClient *sqs.Client
	channel chan types.Message
}
