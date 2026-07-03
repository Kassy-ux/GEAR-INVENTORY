package routes

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"inventory-system/internal/handlers"
)

func RegisterItemRoutes(e *echo.Echo, db *sql.DB) {
	h := handlers.NewItemHandler(db)

	items := e.Group("/items")
	items.POST("", h.CreateItem)
	items.GET("", h.GetItems)
	items.GET("/:id", h.GetItemByID)
	items.PUT("/:id", h.UpdateItem)
	items.DELETE("/:id", h.DeleteItem)
}
