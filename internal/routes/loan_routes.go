package routes

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"inventory-system/internal/handlers"
)

func RegisterLoanRoutes(e *echo.Echo, db *sql.DB) {
	h := handlers.NewLoanHandler(db)

	loans := e.Group("/loans")
	loans.POST("", h.CreateLoan)
	loans.GET("", h.GetLoans)
	loans.GET("/:id", h.GetLoanByID)
	loans.PUT("/:id/return", h.ReturnLoan)
}
