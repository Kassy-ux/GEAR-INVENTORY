package routes

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"inventory-system/internal/handlers"
)

func RegisterBorrowerRoutes(e *echo.Echo, db *sql.DB) {
	h := handlers.NewBorrowerHandler(db)

	borrowers := e.Group("/borrowers")
	borrowers.POST("", h.CreateBorrower)
	borrowers.GET("", h.GetBorrowers)
	borrowers.GET("/:id", h.GetBorrowerByID)
	borrowers.PUT("/:id", h.UpdateBorrower)
	borrowers.DELETE("/:id", h.DeleteBorrower)
}
