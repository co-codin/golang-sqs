package reports

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"go-sqs/store"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ReportBuilder struct {
	reportStore *store.ReportStore
	lozClient   *LozClient
	s3Client    *s3.Client
}

func NewReportBuilder(reportStore *store.ReportStore, lozClient *LozClient, s3Client *s3.Client) *ReportBuilder {
	return &ReportBuilder{
		reportStore: reportStore,
		lozClient:   lozClient,
		s3Client:    s3Client,
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

	resp, err := b.lozClient.GetMonsters()
	if err != nil {
		return nil, fmt.Errorf("failed to get monsters from loz client: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no monsters data returned from loz client")
	}

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	csvWriter := csv.NewWriter(gzipWriter)
	header := []string{"Name", "Id", "Category", "Description", "Image", "Common_Locations", "Drops", "Dlc"}
	if err := csvWriter.Write(header); err != nil {
		return nil, fmt.Errorf("Failed to write csv header: %w", err)
	}

	for _, monster := range resp.Data {
		csvRow := []string{
			monster.Name,
			fmt.Sprintf("%d", monster.Id),
			monster.Category,
			monster.Description,
			monster.Image,
			strings.Join(monster.CommonLocations, ", "),
			strings.Join(monster.Drops, ", "),
			strconv.FormatBool(monster.Dlc),
		}

		if err := csvWriter.Write(csvRow); err != nil {
			return nil, fmt.Errorf("failed to write csv row: %w", err)
		}

		if err := csvWriter.Error(); err != nil {
			return nil, fmt.Errorf("failed to write csv row: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush csv: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}

	return report, nil
}
