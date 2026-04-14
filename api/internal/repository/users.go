package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"stacktrace/internal/models"
)

func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	query := `SELECT id, email, created_at FROM users WHERE email = $1`

	var user models.User
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

func CreateUser(db *sql.DB, email string) (*models.User, error) {
	id := uuid.New()
	query := `INSERT INTO users (id, email) VALUES ($1, $2)`

	_, err := db.Exec(query, id, email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.User{ID: id, Email: email}, nil
}

func GetOrCreateUser(db *sql.DB, email string) (*models.User, error) {
	user, err := GetUserByEmail(db, email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	return CreateUser(db, email)
}
