package routes

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"inventory-system/internal/handlers"
)

func RegisterItemRoutes(r chi.Router, db *sql.DB) {
	h := handlers.NewItemHandler(db)
	r.Route("/items", func(r chi.Router) {
		r.Post("/", h.CreateItem)
		r.Get("/", h.GetItems)
		r.Get("/{id}", h.GetItemByID)
		r.Put("/{id}", h.UpdateItem)
		r.Delete("/{id}", h.DeleteItem)
	})
}
