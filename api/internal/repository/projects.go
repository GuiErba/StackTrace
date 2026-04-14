package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"stacktrace/internal/models"
)

func GetProjectByAPIKey(db *sql.DB, apiKey string) (*models.Project, error) {
	apiKeyHash := hashKey(apiKey)

	query := `
		SELECT id, name, slug, api_key, owner_email, created_at
		FROM projects
		WHERE api_key_hash = $1
	`

	var project models.Project
	err := db.QueryRow(query, apiKeyHash).Scan(
		&project.ID, &project.Name, &project.Slug, &project.APIKey,
		&project.OwnerEmail, &project.CreatedAt,
	)
	if err == nil {
		return &project, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	fallbackQuery := `
		SELECT id, name, slug, api_key, owner_email, created_at
		FROM projects
		WHERE api_key = $1
	`

	err = db.QueryRow(fallbackQuery, apiKey).Scan(
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

func CreateProjectWithUser(db *sql.DB, name, slug string, userID uuid.UUID, apiKeyHash, apiKeyPrefix string) (*models.Project, error) {
	id := uuid.New()
	query := `
		INSERT INTO projects (id, name, slug, user_id, api_key_hash, api_key_prefix, owner_email, api_key)
		VALUES ($1, $2, $3, $4, $5, $6,
			(SELECT email FROM users WHERE id = $4),
			$6
		)
	`
	_, err := db.Exec(query, id, name, slug, userID, apiKeyHash, apiKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	slugPtr := &slug
	return &models.Project{ID: id, Name: name, Slug: slugPtr}, nil
}

func GetProjectsByUserID(db *sql.DB, userID uuid.UUID) ([]models.Project, error) {
	query := `
		SELECT id, name, slug, api_key_prefix, owner_email, created_at
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.APIKeyPrefix, &p.OwnerEmail, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func RotateProjectAPIKey(db *sql.DB, projectID, userID uuid.UUID, newHash, newPrefix string) error {
	query := `
		UPDATE projects
		SET api_key_hash = $1, api_key_prefix = $2
		WHERE id = $3 AND user_id = $4
	`

	result, err := db.Exec(query, newHash, newPrefix, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to rotate API key: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("project not found or access denied")
	}

	return nil
}

func GetProjectByIDAndUser(db *sql.DB, projectID, userID uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, name, slug, api_key_prefix, owner_email, created_at
		FROM projects
		WHERE id = $1 AND user_id = $2
	`

	var project models.Project
	err := db.QueryRow(query, projectID, userID).Scan(
		&project.ID, &project.Name, &project.Slug, &project.APIKeyPrefix,
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

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
