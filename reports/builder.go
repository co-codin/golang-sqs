package reports

import (
	"context"
	"fmt"
	"go-sqs/store"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ReportBuilder struct {
	reportStore *store.ReportStore
	lozClient   *LozClient
	s3Client *s3.Client
}

func NewReportBuilder(reportStore *store.ReportStore, lozClient *LozClient, s3Client *s3.Client) *ReportBuilder {
	return &ReportBuilder{
		reportStore: reportStore,
		lozClient:   lozClient,
		s3Client: s3Client,
	}
}

func (b *ReportBuilder) Build(ctx context.Context, userId uuid.UUID, reportId uuid.UUID) (*store.Report, error) {
	report, err := b.reportStore.ByPrimaryKey(ctx, userId, reportId)
	if err != nil {
		return nil, fmt.Errorf("failed to get report by primary key: %w", err)
	}

	if report.StartedAt != nil {
		return report, nil
	}

	now := time.Now()
	report.StartedAt = &now
	report.CompletedAt = nil
	report.FailedAt = nil
	report.ErrorMessage = nil
	report.DownloadUrl = nil
	report.ExpiresAt = nil
	report.OutputFilePath = nil

	report, err = b.reportStore.Update(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to mark report as started: %w", err)
	}

	// TODO
	// resp, err := b.lozClient.GetMonsters()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get monsters from loz client: %w", err)
	// }

	return report, nil
}