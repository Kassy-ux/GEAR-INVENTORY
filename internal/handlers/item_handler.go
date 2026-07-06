package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"inventory-system/internal/models"
)

type ItemHandler struct {
	DB *sql.DB
}

func NewItemHandler(db *sql.DB) *ItemHandler {
	return &ItemHandler{DB: db}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func isDuplicateSerialError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "serial_number")
}

func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if item.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if item.Status == "" {
		item.Status = "available"
	}
	query := `INSERT INTO items (name, category, description, serial_number, image_url, status) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := h.DB.Exec(query, item.Name, item.Category, item.Description, item.SerialNumber, item.ImageURL, item.Status)
	if err != nil {
		if isDuplicateSerialError(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "an item with this serial number already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	item.ID = int(id)
	writeJSON(w, http.StatusCreated, item)
}

func (h *ItemHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	nameFilter := r.URL.Query().Get("name")
	categoryFilter := r.URL.Query().Get("category")

	query := `SELECT id, name, category, COALESCE(description, ''), serial_number, image_url, status, created_at FROM items`
	conditions := []string{}
	args := []interface{}{}

	if nameFilter != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+nameFilter+"%")
	}
	if categoryFilter != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, categoryFilter)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Description, &item.SerialNumber, &item.ImageURL, &item.Status, &item.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var item models.Item
	query := `SELECT id, name, category, COALESCE(description, ''), serial_number, image_url, status, created_at FROM items WHERE id = ?`
	err = h.DB.QueryRow(query, id).Scan(&item.ID, &item.Name, &item.Category, &item.Description, &item.SerialNumber, &item.ImageURL, &item.Status, &item.CreatedAt)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	query := `UPDATE items SET name = ?, category = ?, description = ?, serial_number = ?, image_url = ?, status = ? WHERE id = ?`
	result, err := h.DB.Exec(query, item.Name, item.Category, item.Description, item.SerialNumber, item.ImageURL, item.Status, id)
	if err != nil {
		if isDuplicateSerialError(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "an item with this serial number already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	item.ID = id
	writeJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	result, err := h.DB.Exec(`DELETE FROM items WHERE id = ?`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
