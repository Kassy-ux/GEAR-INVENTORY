package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// RequireAdmin is an Echo middleware that protects routes so only
// requests with a valid admin JWT can proceed.
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
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

		if claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "admin access required",
			})
		}

		c.Set("adminID", claims.AdminID)
		return next(c)
	}
}