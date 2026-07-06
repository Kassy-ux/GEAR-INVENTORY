// Command seed_admin creates the first admin account directly in the
// database. There is no public signup route by design, so this is the
// only way to create an admin.
//
// Usage:
//
//	go run scripts/seed_admin.go -email admin@example.com -password supersecret
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"inventory-system/internal/config"
	"inventory-system/internal/database"
	"inventory-system/internal/database/queries"
)

func main() {
	email := flag.String("email", "", "admin email address")
	password := flag.String("password", "", "admin password (omit to be prompted securely)")
	flag.Parse()

	cfg := config.Load()

	*email = strings.TrimSpace(*email)
	if *email == "" {
		fmt.Print("Admin email: ")
		fmt.Scanln(email)
		*email = strings.TrimSpace(*email)
	}
	if *email == "" {
		log.Fatal("email is required")
	}

	if *password == "" {
		fmt.Print("Admin password: ")
		bytePw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Fatalf("failed to read password: %v", err)
		}
		*password = string(bytePw)
	}
	if len(*password) < 8 {
		log.Fatal("password must be at least 8 characters")
	}

	conn, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	exists, err := queries.EmailExists(ctx, conn, *email)
	if err != nil {
		log.Fatalf("failed checking existing admin: %v", err)
	}
	if exists {
		log.Fatalf("an admin with email %q already exists", *email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed hashing password: %v", err)
	}

	id, err := queries.CreateAdmin(ctx, conn, *email, string(hash))
	if err != nil {
		log.Fatalf("failed creating admin: %v", err)
	}

	fmt.Printf("admin created successfully: id=%d email=%s\n", id, *email)
}
