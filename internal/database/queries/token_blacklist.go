package queries

import (
	"context"
	"database/sql"
	"time"
)

// RevokeToken stores the full raw JWT string in the blacklist so it can
// no longer be used, even though it hasn't naturally expired yet.
func RevokeToken(ctx context.Context, db *sql.DB, token string, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO token_blacklist (token, expires_at) VALUES (?, ?)`,
		token, expiresAt,
	)
	return err
}

// IsTokenRevoked checks whether a given raw JWT string has been blacklisted.
func IsTokenRevoked(ctx context.Context, db *sql.DB, token string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token = ?)`,
		token,
	).Scan(&exists)
	return exists, err
}