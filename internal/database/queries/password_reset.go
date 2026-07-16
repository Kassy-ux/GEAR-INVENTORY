package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrResetTokenInvalid = errors.New("invalid or expired reset token")

// CreatePasswordReset stores a hashed reset token for the given admin.
func CreatePasswordReset(ctx context.Context, db *sql.DB, adminID int, tokenHash string, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO password_resets (admin_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		adminID, tokenHash, expiresAt,
	)
	return err
}

type PasswordReset struct {
	ID      int
	AdminID int
	Used    bool
}

// GetValidPasswordReset looks up an unused, non-expired reset row by token hash.
func GetValidPasswordReset(ctx context.Context, db *sql.DB, tokenHash string) (*PasswordReset, error) {
	var pr PasswordReset
	err := db.QueryRowContext(ctx,
		`SELECT id, admin_id, used FROM password_resets
		 WHERE token_hash = ? AND used = 0 AND expires_at > UTC_TIMESTAMP()`,
		tokenHash,
	).Scan(&pr.ID, &pr.AdminID, &pr.Used)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrResetTokenInvalid
		}
		return nil, err
	}

	return &pr, nil
}

// MarkPasswordResetUsed prevents a reset token from being used a second time.
func MarkPasswordResetUsed(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, `UPDATE password_resets SET used = 1 WHERE id = ?`, id)
	return err
}

// UpdateAdminPassword sets a new password hash for the given admin.
func UpdateAdminPassword(ctx context.Context, db *sql.DB, adminID int, newHash string) error {
	_, err := db.ExecContext(ctx, `UPDATE admins SET password_hash = ? WHERE id = ?`, newHash, adminID)
	return err
}