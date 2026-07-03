package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

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

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "email and password are required")
			return
		}

		admin, err := queries.GetAdminByEmail(r.Context(), db, req.Email)
		if err != nil {
			if errors.Is(err, queries.ErrAdminNotFound) {
				writeError(w, http.StatusUnauthorized, "invalid credentials")
				return
			}
			writeError(w, http.StatusInternalServerError, "something went wrong")
			return
		}

		if !auth.CheckPassword(req.Password, admin.PasswordHash) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token, err := auth.GenerateToken(admin.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{Token: token})
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
