package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OverviewMetrics struct {
	TotalLogs  int            `json:"total_logs"`
	ErrorCount int            `json:"error_count"`
	WarnCount  int            `json:"warn_count"`
	InfoCount  int            `json:"info_count"`
	LogsPerHour []HourlyCount `json:"logs_per_hour"`
}

type HourlyCount struct {
	Hour  time.Time `json:"hour"`
	Count int       `json:"count"`
}

func GetOverviewMetrics(db *sql.DB, projectID uuid.UUID) (*OverviewMetrics, error) {
	metrics := &OverviewMetrics{}

	countQuery := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE level = 'error') as errors,
			COUNT(*) FILTER (WHERE level = 'warn') as warns,
			COUNT(*) FILTER (WHERE level = 'info') as infos
		FROM logs
		WHERE project_id = $1 AND timestamp >= NOW() - INTERVAL '24 hours'
	`

	err := db.QueryRow(countQuery, projectID).Scan(
		&metrics.TotalLogs, &metrics.ErrorCount,
		&metrics.WarnCount, &metrics.InfoCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	hourlyQuery := `
		SELECT
			date_trunc('hour', timestamp) AS hour,
			COUNT(*) AS count
		FROM logs
		WHERE project_id = $1 AND timestamp >= NOW() - INTERVAL '24 hours'
		GROUP BY hour
		ORDER BY hour ASC
	`

	rows, err := db.Query(hourlyQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query hourly metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var hc HourlyCount
		if err := rows.Scan(&hc.Hour, &hc.Count); err != nil {
			return nil, fmt.Errorf("failed to scan hourly count: %w", err)
		}
		metrics.LogsPerHour = append(metrics.LogsPerHour, hc)
	}

	if metrics.LogsPerHour == nil {
		metrics.LogsPerHour = []HourlyCount{}
	}

	return metrics, nil
}
