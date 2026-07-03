package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Connect opens a MySQL connection pool using database/sql.
// dsn format: user:password@tcp(host:port)/dbname?parseTime=true
func Connect(dsn string) *sql.DB {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("unable to open database connection: %v", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}

	log.Println("connected to database")
	return conn
}