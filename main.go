package main

import (
	"log"

"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"inventory-system/internal/auth"
	"inventory-system/internal/config"
	"inventory-system/internal/database"
	"inventory-system/internal/handlers"
	"inventory-system/internal/routes"
)

func main() {
	cfg := config.Load()
	auth.SetSecret(cfg.JWTSecret)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("connected to database successfully")

	cld, err := database.NewCloudinary(cfg)
	if err != nil {
		log.Fatalf("failed to init cloudinary: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Public auth route
	e.POST("/login", handlers.LoginHandler(db))

	// Admin-only routes
	admin := e.Group("/admin")
	admin.Use(auth.RequireAdmin)
	admin.GET("/dashboard", handlers.DashboardHandler)

	routes.RegisterItemRoutes(e, db)
	routes.RegisterUploadRoutes(e, cld)

	log.Println("starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
