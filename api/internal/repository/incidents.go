package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"stacktrace/internal/models"
)

func CreateIncident(db *sql.DB, incident *models.Incident) error {
	incident.ID = uuid.New()

	query := `
		INSERT INTO incidents (id, project_id, title, description, status, started_at)
		VALUES ($1, $2, $3, $4, 'open', NOW())
	`
	_, err := db.Exec(query, incident.ID, incident.ProjectID, incident.Title, incident.Description)
	return err
}

func GetOpenIncident(db *sql.DB, projectID uuid.UUID) (*models.Incident, error) {
	query := `
		SELECT id, project_id, title, description, status, started_at, resolved_at
		FROM incidents
		WHERE project_id = $1 AND status = 'open'
		ORDER BY started_at DESC
		LIMIT 1
	`

	var incident models.Incident
	err := db.QueryRow(query, projectID).Scan(
		&incident.ID, &incident.ProjectID, &incident.Title,
		&incident.Description, &incident.Status, &incident.StartedAt,
		&incident.ResolvedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query open incident: %w", err)
	}

	return &incident, nil
}

func ResolveIncident(db *sql.DB, incidentID uuid.UUID, projectID uuid.UUID) error {
	query := `
		UPDATE incidents
		SET status = 'resolved', resolved_at = NOW()
		WHERE id = $1 AND project_id = $2 AND status = 'open'
	`
	result, err := db.Exec(query, incidentID, projectID)
	if err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("incident not found or already resolved")
	}

	return nil
}

func ListIncidents(db *sql.DB, projectID uuid.UUID) ([]models.Incident, error) {
	query := `
		SELECT id, project_id, title, description, status, started_at, resolved_at
		FROM incidents
		WHERE project_id = $1
		ORDER BY started_at DESC
		LIMIT 50
	`

	rows, err := db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		err := rows.Scan(
			&inc.ID, &inc.ProjectID, &inc.Title,
			&inc.Description, &inc.Status, &inc.StartedAt,
			&inc.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, inc)
	}

	return incidents, nil
}

func ListRecentIncidents(db *sql.DB, projectID uuid.UUID, limit int) ([]models.Incident, error) {
	query := `
		SELECT id, project_id, title, description, status, started_at, resolved_at
		FROM incidents
		WHERE project_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`

	rows, err := db.Query(query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent incidents: %w", err)
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		err := rows.Scan(
			&inc.ID, &inc.ProjectID, &inc.Title,
			&inc.Description, &inc.Status, &inc.StartedAt,
			&inc.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, inc)
	}

	return incidents, nil
}
