package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"inventory-system/internal/auth"
	"inventory-system/internal/database/queries"
)

// --- Forgot password ---

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordHandler generates a reset token for the given email and
// (for now) logs the reset link to the server console instead of emailing
// it. Swap the log.Printf for a real email send once an email provider
// is wired up.
func ForgotPasswordHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ForgotPasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		if req.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email is required"})
		}

		admin, err := queries.GetAdminByEmail(c.Request().Context(), db, req.Email)
		if err != nil {
			// Always return 200 here, even if the email doesn't exist.
			// This prevents attackers from using this endpoint to figure
			// out which emails are registered admins.
			if errors.Is(err, queries.ErrAdminNotFound) {
				return c.JSON(http.StatusOK, map[string]string{
					"message": "if that email exists, a reset link has been sent",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		}

		rawToken, tokenHash, err := generateResetToken()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate reset token"})
		}

		expiresAt := time.Now().UTC().Add(30 * time.Minute)
		if err := queries.CreatePasswordReset(c.Request().Context(), db, admin.ID, tokenHash, expiresAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create reset token"})
		}

		// TODO: replace with a real email send (e.g. SendGrid, Mailgun).
		// The raw token must only ever be sent to the user directly —
		// never store it, never return it in this response.
		resetLink := "http://localhost:3000/reset-password?token=" + rawToken
		log.Printf("[password reset] email=%s link=%s (expires in 30 min)", admin.Email, resetLink)

		return c.JSON(http.StatusOK, map[string]string{
			"message": "if that email exists, a reset link has been sent",
		})
	}
}

// --- Reset password ---

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func ResetPasswordHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ResetPasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		if req.Token == "" || req.NewPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "token and new_password are required"})
		}
		if len(req.NewPassword) < 8 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		}

		tokenHash := hashToken(req.Token)

		reset, err := queries.GetValidPasswordReset(c.Request().Context(), db, tokenHash)
		if err != nil {
			if errors.Is(err, queries.ErrResetTokenInvalid) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid or expired reset token"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		}

		newHash, err := auth.HashPassword(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		}

		if err := queries.UpdateAdminPassword(c.Request().Context(), db, reset.AdminID, newHash); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update password"})
		}

		if err := queries.MarkPasswordResetUsed(c.Request().Context(), db, reset.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to finalize reset"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "password reset successfully"})
	}
}

// --- Token helpers ---

// generateResetToken creates a random token, returning both the raw
// value (sent to the user) and its SHA-256 hash (stored in the DB).
func generateResetToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	hash = hashToken(raw)
	return raw, hash, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}