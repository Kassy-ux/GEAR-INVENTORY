package auth

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"inventory-system/internal/database/queries"
)

// RequireAdmin is an Echo middleware factory that protects routes so
// only requests with a valid, non-revoked admin JWT can proceed.
func RequireAdmin(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing or malformed authorization header",
				})
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := ParseToken(tokenStr)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid or expired token",
				})
			}

			revoked, err := queries.IsTokenRevoked(c.Request().Context(), db, tokenStr)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to verify token",
				})
			}
			if revoked {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "token has been revoked, please log in again",
				})
			}

			if claims.Role != "admin" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "admin access required",
				})
			}

			c.Set("adminID", claims.AdminID)
			return next(c)
		}
	}
}