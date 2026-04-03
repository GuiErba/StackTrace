package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID        int64           `json:"id"`
	ProjectID uuid.UUID       `json:"project_id"`
	Timestamp time.Time       `json:"timestamp"`
	Level     string          `json:"level"`
	Message   string          `json:"message"`
	Service   string          `json:"service"`
	TraceID   *string         `json:"trace_id,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

type LogInput struct {
	Level     string          `json:"level" binding:"required,oneof=info warn error"`
	Message   string          `json:"message" binding:"required"`
	Service   string          `json:"service"`
	Timestamp *time.Time      `json:"timestamp,omitempty"`
	TraceID   *string         `json:"trace_id,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

type LogFilter struct {
	Level   string
	Service string
	TraceID string
	From    *time.Time
	To      *time.Time
	Cursor  string
	Limit   int
}
