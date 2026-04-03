package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	APIKey     string    `json:"api_key"`
	OwnerEmail string    `json:"owner_email"`
	CreatedAt  time.Time `json:"created_at"`
}
