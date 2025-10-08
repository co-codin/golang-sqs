package store

import "database/sql"

type Store struct {
	Users *UserStore
	RefreshTokens *RefreshTokenStore
	ReportStore *ReportStore
}

func New(db *sql.DB) *Store {
	return &Store{
		Users: NewUserStore(db),
		RefreshTokens: NewRefreshTokenStore(db),
		ReportStore: NewReportStore(db),
	}
}
