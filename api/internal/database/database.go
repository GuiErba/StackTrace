package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	if strings.Contains(databaseURL, "pooler") && !strings.Contains(databaseURL, "prefer_simple_protocol") {
		separator := "&"
		if !strings.Contains(databaseURL, "?") {
			separator = "?"
		}
		databaseURL = databaseURL + separator + "prefer_simple_protocol=true"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
