package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Slug         *string    `json:"slug,omitempty"`
	APIKey       string     `json:"-"`
	APIKeyPrefix *string    `json:"api_key_prefix,omitempty"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	OwnerEmail   string     `json:"owner_email"`
	CreatedAt    time.Time  `json:"created_at"`
}
