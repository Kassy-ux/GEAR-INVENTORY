package database

import (
	"context"
	"database/sql"
	"errors"
)

var ErrAdminNotFound = errors.New("admin not found")

type Admin struct {
	ID           int
	Email        string
	
	PasswordHash string
}

// GetAdminByEmail fetches a single admin row by email.
// Used by the login handler to verify credentials.
func GetAdminByEmail(ctx context.Context, db *sql.DB, email string) (*Admin, error) {
	var a Admin
	err := db.QueryRowContext(ctx,
		`SELECT id, email, password_hash FROM admins WHERE email = ?`,
		email,
	).Scan(&a.ID, &a.Email, &a.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}

	return &a, nil
}