package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"inventory-system/internal/auth"
	"inventory-system/internal/config"
	"inventory-system/internal/database"
	"inventory-system/internal/handlers"
	"inventory-system/internal/routes"
)

func main() {
	cfg := config.Load()
	auth.SetSecret(cfg.JWTSecret)

	conn := database.Connect(cfg.DatabaseDSN())
	defer conn.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Post("/login", handlers.LoginHandler(conn))
	routes.RegisterItemRoutes(r, conn)

	// Admin-only routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAdmin)
		r.Get("/admin/dashboard", handlers.DashboardHandler)
	})

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}