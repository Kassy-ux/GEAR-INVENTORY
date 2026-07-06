package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// DashboardHandler is an example route protected by auth.RequireAdmin.
func DashboardHandler(c echo.Context) error {
	adminID, _ := c.Get("adminID").(string)

	return c.JSON(http.StatusOK, map[string]string{
		"message":  "welcome to the admin dashboard",
		"admin_id": adminID,
	})
}