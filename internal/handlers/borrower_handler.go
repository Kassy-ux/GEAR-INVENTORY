package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"inventory-system/internal/models"
)

type BorrowerHandler struct {
	DB *sql.DB
}

func NewBorrowerHandler(db *sql.DB) *BorrowerHandler {
	return &BorrowerHandler{DB: db}
}

func isDuplicateEmailError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "email")
}

// POST /borrowers
func (h *BorrowerHandler) CreateBorrower(c echo.Context) error {
	var b models.Borrower
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if b.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	result, err := h.DB.Exec(`INSERT INTO borrowers (name, email, department) VALUES (?, ?, ?)`, b.Name, b.Email, b.Department)
	if err != nil {
		if isDuplicateEmailError(err) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "a borrower with this email already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	id, _ := result.LastInsertId()
	b.ID = int(id)

	return c.JSON(http.StatusCreated, b)
}

// GET /borrowers?name=jane
func (h *BorrowerHandler) GetBorrowers(c echo.Context) error {
	nameFilter := c.QueryParam("name")

	query := `SELECT id, name, COALESCE(email, ''), COALESCE(department, '') FROM borrowers`
	args := []interface{}{}
	if nameFilter != "" {
		query += ` WHERE name LIKE ?`
		args = append(args, "%"+nameFilter+"%")
	}

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	borrowers := []models.Borrower{}
	for rows.Next() {
		var b models.Borrower
		if err := rows.Scan(&b.ID, &b.Name, &b.Email, &b.Department); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		borrowers = append(borrowers, b)
	}

	return c.JSON(http.StatusOK, borrowers)
}

// GET /borrowers/:id
func (h *BorrowerHandler) GetBorrowerByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var b models.Borrower
	err = h.DB.QueryRow(`SELECT id, name, COALESCE(email, ''), COALESCE(department, '') FROM borrowers WHERE id = ?`, id).
		Scan(&b.ID, &b.Name, &b.Email, &b.Department)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "borrower not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, b)
}

// PUT /borrowers/:id
func (h *BorrowerHandler) UpdateBorrower(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var b models.Borrower
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	result, err := h.DB.Exec(`UPDATE borrowers SET name = ?, email = ?, department = ? WHERE id = ?`, b.Name, b.Email, b.Department, id)
	if err != nil {
		if isDuplicateEmailError(err) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "a borrower with this email already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "borrower not found"})
	}

	b.ID = id
	return c.JSON(http.StatusOK, b)
}

// DELETE /borrowers/:id
func (h *BorrowerHandler) DeleteBorrower(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	result, err := h.DB.Exec(`DELETE FROM borrowers WHERE id = ?`, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "borrower not found"})
	}

	return c.JSON(http.StatusNoContent, nil)
}
