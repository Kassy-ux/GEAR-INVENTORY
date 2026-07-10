package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"inventory-system/internal/models"
)

type LoanHandler struct {
	DB *sql.DB
}

func NewLoanHandler(db *sql.DB) *LoanHandler {
	return &LoanHandler{DB: db}
}

type CreateLoanRequest struct {
	ItemID     int    `json:"item_id"`
	BorrowerID int    `json:"borrower_id"`
	DueDate    string `json:"due_date"` // format: "2026-07-20"
}

// POST /loans  (checkout)
func (h *LoanHandler) CreateLoan(c echo.Context) error {
	var req CreateLoanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.ItemID == 0 || req.BorrowerID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "item_id and borrower_id are required"})
	}

	tx, err := h.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer tx.Rollback()

	// 1. Check item status
	var status string
	err = tx.QueryRow(`SELECT status FROM items WHERE id = ?`, req.ItemID).Scan(&status)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if status != "available" {
		return c.JSON(http.StatusConflict, map[string]string{"error": "item is not available for checkout"})
	}

	// 2. Confirm borrower exists
	var borrowerExists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM borrowers WHERE id = ?)`, req.BorrowerID).Scan(&borrowerExists)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if !borrowerExists {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "borrower not found"})
	}

	// 3. Create the loan
	var dueDate interface{}
	if req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "due_date must be in YYYY-MM-DD format"})
		}
		dueDate = parsed
	}

	result, err := tx.Exec(`INSERT INTO loans (item_id, borrower_id, checked_out_at, due_date) VALUES (?, ?, NOW(), ?)`,
		req.ItemID, req.BorrowerID, dueDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	loanID, _ := result.LastInsertId()

	// 4. Update item status
	_, err = tx.Exec(`UPDATE items SET status = 'checked_out' WHERE id = ?`, req.ItemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":          loanID,
		"item_id":     req.ItemID,
		"borrower_id": req.BorrowerID,
		"message":     "item checked out successfully",
	})
}

// PUT /loans/:id/return
func (h *LoanHandler) ReturnLoan(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	tx, err := h.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer tx.Rollback()

	var itemID int
	var returnedAt sql.NullTime
	err = tx.QueryRow(`SELECT item_id, returned_at FROM loans WHERE id = ?`, id).Scan(&itemID, &returnedAt)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "loan not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if returnedAt.Valid {
		return c.JSON(http.StatusConflict, map[string]string{"error": "this loan has already been returned"})
	}

	_, err = tx.Exec(`UPDATE loans SET returned_at = NOW() WHERE id = ?`, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	_, err = tx.Exec(`UPDATE items SET status = 'available' WHERE id = ?`, itemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "item returned successfully"})
}

// GET /loans?active=true
func (h *LoanHandler) GetLoans(c echo.Context) error {
	activeOnly := c.QueryParam("active")

	query := `SELECT id, item_id, borrower_id, checked_out_at, due_date, returned_at FROM loans`
	if activeOnly == "true" {
		query += ` WHERE returned_at IS NULL`
	}

	rows, err := h.DB.Query(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	loans := []models.Loan{}
	for rows.Next() {
		var l models.Loan
		if err := rows.Scan(&l.ID, &l.ItemID, &l.BorrowerID, &l.CheckedOutAt, &l.DueDate, &l.ReturnedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		loans = append(loans, l)
	}

	return c.JSON(http.StatusOK, loans)
}

// GET /loans/:id
func (h *LoanHandler) GetLoanByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var l models.Loan
	err = h.DB.QueryRow(`SELECT id, item_id, borrower_id, checked_out_at, due_date, returned_at FROM loans WHERE id = ?`, id).
		Scan(&l.ID, &l.ItemID, &l.BorrowerID, &l.CheckedOutAt, &l.DueDate, &l.ReturnedAt)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "loan not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, l)
}
