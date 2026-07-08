package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"inventory-system/internal/auth"
	"inventory-system/internal/database/queries"
)

// LogoutHandler revokes the token used to authenticate this request,
// so it can no longer be used even though it hasn't expired yet.
// Must sit behind auth.RequireAdmin so the token is already validated.
func LogoutHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
		}

		err = queries.RevokeToken(c.Request().Context(), db, tokenStr, claims.ExpiresAt.Time)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to log out"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "logged out successfully"})
	}
}