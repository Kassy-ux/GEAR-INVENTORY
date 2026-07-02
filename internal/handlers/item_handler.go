package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"inventory-system/internal/models"
)

type ItemHandler struct {
	DB *sql.DB
}

func NewItemHandler(db *sql.DB) *ItemHandler {
	return &ItemHandler{DB: db}
}

// POST /items
func (h *ItemHandler) CreateItem(c echo.Context) error {
	var item models.Item
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if item.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	if item.Status == "" {
		item.Status = "available"
	}

	query := `INSERT INTO items (name, category, serial_number, image_url, status) VALUES (?, ?, ?, ?, ?)`
	result, err := h.DB.Exec(query, item.Name, item.Category, item.SerialNumber, item.ImageURL, item.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	id, _ := result.LastInsertId()
	item.ID = int(id)

	return c.JSON(http.StatusCreated, item)
}

// GET /items?name=laptop
func (h *ItemHandler) GetItems(c echo.Context) error {
	nameFilter := c.QueryParam("name")

	query := `SELECT id, name, category, serial_number, image_url, status, created_at FROM items`
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

	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.SerialNumber, &item.ImageURL, &item.Status, &item.CreatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		items = append(items, item)
	}

	return c.JSON(http.StatusOK, items)
}

// GET /items/:id
func (h *ItemHandler) GetItemByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var item models.Item
	query := `SELECT id, name, category, serial_number, image_url, status, created_at FROM items WHERE id = ?`
	err = h.DB.QueryRow(query, id).Scan(&item.ID, &item.Name, &item.Category, &item.SerialNumber, &item.ImageURL, &item.Status, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, item)
}

// PUT /items/:id
func (h *ItemHandler) UpdateItem(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var item models.Item
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	query := `UPDATE items SET name = ?, category = ?, serial_number = ?, image_url = ?, status = ? WHERE id = ?`
	result, err := h.DB.Exec(query, item.Name, item.Category, item.SerialNumber, item.ImageURL, item.Status, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}

	item.ID = id
	return c.JSON(http.StatusOK, item)
}

// DELETE /items/:id
func (h *ItemHandler) DeleteItem(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	result, err := h.DB.Exec(`DELETE FROM items WHERE id = ?`, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}

	return c.JSON(http.StatusNoContent, nil)
}
