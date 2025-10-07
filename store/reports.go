package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ReportStore struct {
	db *sqlx.DB
}

func NewReportStore(db *sql.DB) *ReportStore {
	return &ReportStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type Report struct {
	UserID         uuid.UUID  `db:"user_id"`
	Id             uuid.UUID  `db:"id"`
	ReportType     string     `db:"report_type"`
	OutputFilePath *string    `db:"output_file_path"`
	DownloadUrl    *string    `db:"download_url"`
	ExpiresAt      *time.Time `db:"expires_at"`
	ErrorMessage   *string    `db:"error_message"`
	CreatedAt      time.Time  `db:"created_at"`
	StartedAt      time.Time  `db:"started_at"`
	FailedAt       *time.Time `db:"failed_at"`
	CompletedAt    *time.Time `db:"completed_at"`
}

