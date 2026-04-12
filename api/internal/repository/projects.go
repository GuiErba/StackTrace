package repository

import (
	"database/sql"
	"fmt"

	"stacktrace/internal/models"
)

func GetProjectByAPIKey(db *sql.DB, apiKey string) (*models.Project, error) {
	query := `
		SELECT id, name, slug, api_key, owner_email, created_at
		FROM projects
		WHERE api_key = $1
	`

	var project models.Project
	err := db.QueryRow(query, apiKey).Scan(
		&project.ID, &project.Name, &project.Slug, &project.APIKey,
		&project.OwnerEmail, &project.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	return &project, nil
}

func GetProjectBySlug(db *sql.DB, slug string) (*models.Project, error) {
	query := `
		SELECT id, name, slug, api_key, owner_email, created_at
		FROM projects
		WHERE slug = $1
	`

	var project models.Project
	err := db.QueryRow(query, slug).Scan(
		&project.ID, &project.Name, &project.Slug, &project.APIKey,
		&project.OwnerEmail, &project.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	return &project, nil
}
