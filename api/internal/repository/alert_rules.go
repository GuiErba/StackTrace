package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"stacktrace/internal/models"
)

func CreateAlertRule(db *sql.DB, rule *models.AlertRule) error {
	rule.ID = uuid.New()

	query := `
		INSERT INTO alert_rules (id, project_id, condition, threshold, window_seconds, channel, destination, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := db.Exec(query,
		rule.ID, rule.ProjectID, rule.Condition, rule.Threshold,
		rule.WindowSeconds, rule.Channel, rule.Destination,
	)
	return err
}

func ListAlertRules(db *sql.DB, projectID uuid.UUID) ([]models.AlertRule, error) {
	query := `
		SELECT id, project_id, condition, threshold, window_seconds, channel, destination, created_at
		FROM alert_rules
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert rules: %w", err)
	}
	defer rows.Close()

	var rules []models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		err := rows.Scan(
			&rule.ID, &rule.ProjectID, &rule.Condition, &rule.Threshold,
			&rule.WindowSeconds, &rule.Channel, &rule.Destination, &rule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func DeleteAlertRule(db *sql.DB, ruleID uuid.UUID, projectID uuid.UUID) error {
	query := `DELETE FROM alert_rules WHERE id = $1 AND project_id = $2`

	result, err := db.Exec(query, ruleID, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("alert rule not found")
	}

	return nil
}

func GetAlertRulesByProjectID(db *sql.DB, projectID uuid.UUID) ([]models.AlertRule, error) {
	return ListAlertRules(db, projectID)
}
