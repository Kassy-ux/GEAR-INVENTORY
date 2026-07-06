package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"inventory-system/internal/config"
)

// Connect opens a MySQL connection pool using database/sql.
func Connect(cfg *config.Config) (*sql.DB, error) {
	conn, err := sql.Open("mysql", cfg.DatabaseDSN())
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	log.Println("connected to database")
	return conn, nil
}