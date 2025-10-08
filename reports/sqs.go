package reports

import "github.com/google/uuid"

type SqsMessage struct {
	userId   uuid.UUID `json:"userId"`
	ReportId uuid.UUID `json:"reportId"`
}
