package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const AdminIDKey contextKey = "adminID"

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeJSONError(w, http.StatusUnauthorized, "missing or malformed authorization header")
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ParseToken(tokenStr)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		if claims.Role != "admin" {
			writeJSONError(w, http.StatusForbidden, "admin access required")
			return
		}

		ctx := context.WithValue(r.Context(), AdminIDKey, claims.AdminID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}