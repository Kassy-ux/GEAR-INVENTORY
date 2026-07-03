package queries

import (
	"context"
	"database/sql"
)

// CreateAdmin inserts a new admin row and returns the generated ID.
// Used only by scripts/seed_admin.go — there is no public signup route.
func CreateAdmin(ctx context.Context, db *sql.DB, email, passwordHash string) (int, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO admins (email, password_hash) VALUES (?, ?)`,
		email, passwordHash,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// EmailExists checks whether an admin with the given email already exists.
// Used only by scripts/seed_admin.go to avoid duplicate admins.
func EmailExists(ctx context.Context, db *sql.DB, email string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM admins WHERE email = ?)`,
		email,
	).Scan(&exists)
	return exists, err
}
