package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"inventory-system/internal/auth"
	"inventory-system/internal/database/queries"
)

// --- Forgot password ---

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordHandler generates a reset token for the given email.
//
// There is no real email service wired up yet. In development (appEnv
// != "production"), the raw token is returned directly in the JSON
// response and logged to the console, so the frontend can test the
// full reset flow without needing the backend dev to manually relay
// the token. In production, the token is ONLY ever logged/emailed —
// never returned in the API response, since that would let anyone
// reset any admin's password just by knowing their email.
func ForgotPasswordHandler(db *sql.DB, appEnv string) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ForgotPasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		if req.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email is required"})
		}

		genericResponse := map[string]string{
			"message": "if that email exists, a reset link has been sent",
		}

		admin, err := queries.GetAdminByEmail(c.Request().Context(), db, req.Email)
		if err != nil {
			// Always return 200 here, even if the email doesn't exist.
			// This prevents attackers from using this endpoint to figure
			// out which emails are registered admins.
			if errors.Is(err, queries.ErrAdminNotFound) {
				return c.JSON(http.StatusOK, genericResponse)
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

		resetLink := "http://localhost:3000/reset-password?token=" + rawToken
		log.Printf("[password reset] email=%s link=%s (expires in 30 min)", admin.Email, resetLink)

		// DEV-ONLY: return the token directly so the frontend can test
		// the full flow without a real email service. This branch must
		// never run in production — it would let anyone reset any
		// admin's password just by knowing their email address.
		if appEnv != "production" {
			return c.JSON(http.StatusOK, map[string]string{
				"message":    "if that email exists, a reset link has been sent",
				"dev_token":  rawToken,
				"dev_notice": "this field only appears when APP_ENV is not 'production'",
			})
		}

		return c.JSON(http.StatusOK, genericResponse)
	}
}

// --- Reset password ---

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// ResetPasswordHandler expects the reset token as a Bearer token in the
// Authorization header, exactly like every other authenticated route in
// this API — not as a field in the JSON body. This keeps the token
// handling consistent across the whole API and means the frontend never
// needs a special case just for this one endpoint: whatever mechanism
// already attaches a Bearer token to a request works here unchanged.
func ResetPasswordHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid Authorization header"})
		}
		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		var req ResetPasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		if req.NewPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "new_password is required"})
		}
		if len(req.NewPassword) < 8 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		}

		tokenHash := hashToken(rawToken)

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