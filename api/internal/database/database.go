package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) (*sql.DB, error) {
	if strings.Contains(databaseURL, "pooler") {
		separator := "&"
		if !strings.Contains(databaseURL, "?") {
			separator = "?"
		}
		// For pgx with PgBouncer, force simple query execution to avoid prepared statement drops
		if !strings.Contains(databaseURL, "default_query_exec_mode") {
			databaseURL = databaseURL + separator + "default_query_exec_mode=exec"
		}
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
