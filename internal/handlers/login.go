package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"inventory-system/internal/auth"
	"inventory-system/internal/database/queries"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// LoginHandler authenticates an admin by email/password and returns a JWT.
func LoginHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		if req.Email == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		}

		admin, err := queries.GetAdminByEmail(c.Request().Context(), db, req.Email)
		if err != nil {
			if errors.Is(err, queries.ErrAdminNotFound) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		}

		if !auth.CheckPassword(req.Password, admin.PasswordHash) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}

		token, err := auth.GenerateToken(admin.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}

		return c.JSON(http.StatusOK, LoginResponse{Token: token})
	}
}