package handlers

import (
	"encoding/json"
	"net/http"

	"inventory-system/internal/auth"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(auth.AdminIDKey).(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "welcome to the admin dashboard",
		"admin_id": adminID,
	})
}
