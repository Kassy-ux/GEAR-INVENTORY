package queries

import (
	"context"
	"database/sql"
)

type OverdueLoan struct {
	LoanID       int
	ItemName     string
	BorrowerName string
	DueDate      string
}

// GetUnnotifiedOverdueLoans finds loans that are past their due_date,
// not yet returned, and haven't already triggered a notification.
func GetUnnotifiedOverdueLoans(ctx context.Context, db *sql.DB) ([]OverdueLoan, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT l.id, i.name, b.name, l.due_date
		FROM loans l
		JOIN items i ON i.id = l.item_id
		JOIN borrowers b ON b.id = l.borrower_id
		WHERE l.due_date < NOW()
		  AND l.returned_at IS NULL
		  AND l.notified = 0
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overdue []OverdueLoan
	for rows.Next() {
		var o OverdueLoan
		if err := rows.Scan(&o.LoanID, &o.ItemName, &o.BorrowerName, &o.DueDate); err != nil {
			return nil, err
		}
		overdue = append(overdue, o)
	}
	return overdue, nil
}

// MarkLoanNotified flags a loan so it won't be picked up again by the checker.
func MarkLoanNotified(ctx context.Context, db *sql.DB, loanID int) error {
	_, err := db.ExecContext(ctx, `UPDATE loans SET notified = 1 WHERE id = ?`, loanID)
	return err
}

// CreateNotification inserts a new notification for an overdue loan.
func CreateNotification(ctx context.Context, db *sql.DB, loanID int, message string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO notifications (loan_id, message) VALUES (?, ?)`,
		loanID, message,
	)
	return err
}

type Notification struct {
	ID        int    `json:"id"`
	LoanID    int    `json:"loan_id"`
	Message   string `json:"message"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

// ListNotifications returns notifications, most recent first.
// If unreadOnly is true, only unread notifications are returned.
func ListNotifications(ctx context.Context, db *sql.DB, unreadOnly bool) ([]Notification, error) {
	query := `SELECT id, loan_id, message, is_read, created_at FROM notifications`
	if unreadOnly {
		query += ` WHERE is_read = 0`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.LoanID, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

// MarkNotificationRead marks a single notification as read.
func MarkNotificationRead(ctx context.Context, db *sql.DB, id int) (bool, error) {
	result, err := db.ExecContext(ctx, `UPDATE notifications SET is_read = 1 WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}