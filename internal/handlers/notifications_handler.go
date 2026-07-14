package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"inventory-system/internal/database/queries"
)

// ListNotificationsHandler returns notifications, most recent first.
// Supports ?unread=true to filter to only unread ones.
func ListNotificationsHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		unreadOnly := c.QueryParam("unread") == "true"

		notifications, err := queries.ListNotifications(c.Request().Context(), db, unreadOnly)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, notifications)
	}
}

// MarkNotificationReadHandler marks a single notification as read.
func MarkNotificationReadHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
		}

		found, err := queries.MarkNotificationRead(c.Request().Context(), db, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if !found {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "notification not found"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "marked as read"})
	}
}