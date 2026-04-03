package repository

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"stacktrace/internal/models"
)

func InsertLog(db *sql.DB, log *models.LogEntry) error {
	query := `
		INSERT INTO logs (project_id, timestamp, level, message, service, trace_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(query,
		log.ProjectID, log.Timestamp, log.Level, log.Message,
		log.Service, log.TraceID, log.Metadata,
	)
	return err
}

func InsertLogBatch(db *sql.DB, logs []models.LogEntry) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO logs (project_id, timestamp, level, message, service, trace_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, log := range logs {
		_, err := stmt.Exec(
			log.ProjectID, log.Timestamp, log.Level, log.Message,
			log.Service, log.TraceID, log.Metadata,
		)
		if err != nil {
			return fmt.Errorf("failed to insert log: %w", err)
		}
	}

	return tx.Commit()
}

func QueryLogs(db *sql.DB, projectID uuid.UUID, filters models.LogFilter) ([]models.LogEntry, string, bool, error) {
	limit := filters.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	queryLimit := limit + 1

	query := `SELECT id, project_id, timestamp, level, message, service, trace_id, metadata FROM logs WHERE project_id = $1`
	args := []interface{}{projectID}
	argIndex := 2

	if filters.Level != "" {
		query += fmt.Sprintf(" AND level = $%d", argIndex)
		args = append(args, filters.Level)
		argIndex++
	}

	if filters.Service != "" {
		query += fmt.Sprintf(" AND service = $%d", argIndex)
		args = append(args, filters.Service)
		argIndex++
	}

	if filters.TraceID != "" {
		query += fmt.Sprintf(" AND trace_id = $%d", argIndex)
		args = append(args, filters.TraceID)
		argIndex++
	}

	if filters.From != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filters.From)
		argIndex++
	}

	if filters.To != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filters.To)
		argIndex++
	}

	if filters.Cursor != "" {
		cursorID, cursorTS, err := decodeCursor(filters.Cursor)
		if err == nil {
			query += fmt.Sprintf(" AND (timestamp, id) < ($%d, $%d)", argIndex, argIndex+1)
			args = append(args, cursorTS, cursorID)
			argIndex += 2
		}
	}

	query += fmt.Sprintf(" ORDER BY timestamp DESC, id DESC LIMIT $%d", argIndex)
	args = append(args, queryLimit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var log models.LogEntry
		var metadata sql.NullString
		var traceID sql.NullString

		err := rows.Scan(
			&log.ID, &log.ProjectID, &log.Timestamp, &log.Level,
			&log.Message, &log.Service, &traceID, &metadata,
		)
		if err != nil {
			return nil, "", false, fmt.Errorf("failed to scan log: %w", err)
		}

		if traceID.Valid {
			log.TraceID = &traceID.String
		}
		if metadata.Valid {
			log.Metadata = json.RawMessage(metadata.String)
		}

		logs = append(logs, log)
	}

	hasMore := len(logs) > limit
	if hasMore {
		logs = logs[:limit]
	}

	var nextCursor string
	if hasMore && len(logs) > 0 {
		last := logs[len(logs)-1]
		nextCursor = encodeCursor(last.ID, last.Timestamp)
	}

	return logs, nextCursor, hasMore, nil
}

func encodeCursor(id int64, ts time.Time) string {
	raw := fmt.Sprintf("%d,%s", id, ts.Format(time.RFC3339Nano))
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (int64, time.Time, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, time.Time{}, err
	}

	parts := strings.SplitN(string(decoded), ",", 2)
	if len(parts) != 2 {
		return 0, time.Time{}, fmt.Errorf("invalid cursor format")
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, time.Time{}, err
	}

	ts, err := time.Parse(time.RFC3339Nano, parts[1])
	if err != nil {
		return 0, time.Time{}, err
	}

	return id, ts, nil
}
