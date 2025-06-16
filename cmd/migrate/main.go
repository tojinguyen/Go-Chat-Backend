package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gochat-backend/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if not in release mode
	if os.Getenv("RUN_MODE") != "release" {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Warning: Failed to load .env file: %v", err)
		}
	}

	// Load environment configuration
	env, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}

	// Create database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		env.MysqlUser,
		env.MysqlPassword,
		env.MysqlHost,
		env.MysqlPort,
		env.MysqlDatabase,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Create migrations table if not exists
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(createMigrationsTable); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get migration mode from environment or default to "up"
	migrateMode := env.MysqlMigrateMode
	if migrateMode == "" || migrateMode == "auto" {
		migrateMode = "up"
	}

	if migrateMode == "up" {
		if err := runMigrationsUp(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	} else if migrateMode == "down" {
		if err := runMigrationsDown(db); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
	}

	log.Println("Migrations completed successfully")
}

func runMigrationsUp(db *sql.DB) error {
	// Get all migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Run unapplied migrations
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; !applied {
			log.Printf("Running migration: %s", file)

			if err := runMigrationFile(db, file); err != nil {
				return fmt.Errorf("failed to run migration %s: %v", file, err)
			}

			// Record migration as applied
			if err := recordMigration(db, file); err != nil {
				return fmt.Errorf("failed to record migration %s: %v", file, err)
			}

			log.Printf("Migration %s completed successfully", file)
		}
	}

	return nil
}

func runMigrationsDown(db *sql.DB) error {
	// Get all migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Reverse the order for rollback
	sort.Sort(sort.Reverse(sort.StringSlice(migrationFiles)))

	// Run rollback for applied migrations
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; applied {
			log.Printf("Rolling back migration: %s", file)

			// Check if down migration file exists
			downFile := strings.Replace(file, ".sql", ".down.sql", 1)
			if err := runMigrationFile(db, downFile); err != nil {
				log.Printf("Warning: Failed to run down migration %s: %v", downFile, err)
				continue
			}

			// Remove migration record
			if err := removeMigrationRecord(db, file); err != nil {
				return fmt.Errorf("failed to remove migration record %s: %v", file, err)
			}

			log.Printf("Migration %s rolled back successfully", file)
		}
	}

	return nil
}

func getMigrationFiles() ([]string, error) {
	var files []string
	migrationDir := "migrations/mysql"

	err := filepath.WalkDir(migrationDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".sql") && !strings.HasSuffix(path, ".down.sql") {
			files = append(files, filepath.Base(path))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := db.Query("SELECT filename FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, nil
}

func runMigrationFile(db *sql.DB, filename string) error {
	migrationPath := filepath.Join("migrations/mysql", filename)

	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}

	// Split content by semicolon and execute each statement
	statements := strings.Split(string(content), ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %s, error: %v", stmt, err)
		}
	}

	return nil
}

func recordMigration(db *sql.DB, filename string) error {
	_, err := db.Exec("INSERT INTO migrations (filename) VALUES (?)", filename)
	return err
}

func removeMigrationRecord(db *sql.DB, filename string) error {
	_, err := db.Exec("DELETE FROM migrations WHERE filename = ?", filename)
	return err
}
