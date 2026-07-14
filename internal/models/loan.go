package models

import "time"

type Loan struct {
	ID           int        `json:"id"`
	ItemID       int        `json:"item_id"`
	BorrowerID   int        `json:"borrower_id"`
	CheckedOutAt time.Time  `json:"checked_out_at"`
	DueDate      *time.Time `json:"due_date"`
	ReturnedAt   *time.Time `json:"returned_at"`
}

type LoanWithDetails struct {
	ID           int        `json:"id"`
	ItemID       int        `json:"item_id"`
	ItemName     string     `json:"item_name"`
	BorrowerID   int        `json:"borrower_id"`
	BorrowerName string     `json:"borrower_name"`
	CheckedOutAt time.Time  `json:"checked_out_at"`
	DueDate      *time.Time `json:"due_date"`
	ReturnedAt   *time.Time `json:"returned_at"`
}
