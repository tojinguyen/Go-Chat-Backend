package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")

	db, err := sql.Open("mysql", dbURL)

	if err != nil {
		log.Fatal("Database Connection Failed:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database Ping Failed:", err)
	}

	return db
}