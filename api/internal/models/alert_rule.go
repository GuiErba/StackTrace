package models

import (
	"time"

	"github.com/google/uuid"
)

type AlertRule struct {
	ID            uuid.UUID `json:"id"`
	ProjectID     uuid.UUID `json:"project_id"`
	Condition     string    `json:"condition"`
	Threshold     int       `json:"threshold"`
	WindowSeconds int       `json:"window_seconds"`
	Channel       string    `json:"channel"`
	Destination   string    `json:"destination"`
	CreatedAt     time.Time `json:"created_at"`
}
