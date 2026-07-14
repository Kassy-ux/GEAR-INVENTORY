package notifier

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"inventory-system/internal/database/queries"
)

// StartOverdueChecker runs a background loop that periodically scans
// for loans past their due date and creates a notification for each
// one found, exactly once per loan (guarded by the loans.notified flag).
//
// Call this once from main.go with `go notifier.StartOverdueChecker(db, 10*time.Minute)`.
// The interval controls how often the check runs — 10-15 minutes is a
// reasonable default for an inventory system; tighten it if admins need
// faster alerts.
func StartOverdueChecker(db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run once immediately on startup, then on every tick.
	checkOverdueLoans(db)

	for range ticker.C {
		checkOverdueLoans(db)
	}
}

func checkOverdueLoans(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	overdue, err := queries.GetUnnotifiedOverdueLoans(ctx, db)
	if err != nil {
		log.Printf("[notifier] failed to check overdue loans: %v", err)
		return
	}

	for _, loan := range overdue {
		message := fmt.Sprintf(
			"Item %q borrowed by %s is overdue (was due %s).",
			loan.ItemName, loan.BorrowerName, loan.DueDate,
		)

		if err := queries.CreateNotification(ctx, db, loan.LoanID, message); err != nil {
			log.Printf("[notifier] failed to create notification for loan %d: %v", loan.LoanID, err)
			continue
		}

		if err := queries.MarkLoanNotified(ctx, db, loan.LoanID); err != nil {
			log.Printf("[notifier] failed to mark loan %d as notified: %v", loan.LoanID, err)
			continue
		}

		log.Printf("[notifier] created overdue notification for loan %d", loan.LoanID)
	}
}