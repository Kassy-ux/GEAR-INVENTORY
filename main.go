package main

import (
	"log"

	"inventory-system/internal/auth"
	"inventory-system/internal/config"
	"inventory-system/internal/database"
	"inventory-system/internal/handlers"
	"inventory-system/internal/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// Public auth routes
	e.POST("/login", handlers.LoginHandler(db))
	e.POST("/forgot-password", handlers.ForgotPasswordHandler(db))
	e.POST("/reset-password", handlers.ResetPasswordHandler(db))

	// Admin-only routes
	admin := e.Group("/admin")
	admin.Use(auth.RequireAdmin(db))
	admin.GET("/dashboard", handlers.DashboardHandler)

	// Logout — requires a valid token, but isn't nested under /admin
	e.POST("/logout", handlers.LogoutHandler(db), auth.RequireAdmin(db))

	routes.RegisterItemRoutes(e, db)
	routes.RegisterUploadRoutes(e, cld)

	log.Println("starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}

