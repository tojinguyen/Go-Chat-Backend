package main

import (
	"fmt"
	"gochat-backend/internal/config"
	"gochat-backend/internal/infra/mysqlinfra"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Xác định hành động migration từ tham số dòng lệnh
	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := mysqlinfra.ConnectMysql(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	database := mysqlinfra.NewMySqlDatabase(db)
	defer database.Close()

	// Run migrations
	if err := runMigration(database, action); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migrations completed successfully")
}

func runMigration(db *mysqlinfra.Database, action string) error {
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/mysql",
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	switch action {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "reset":
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to rollback migrations: %w", err)
		}
		err = m.Up()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration %s failed: %w", action, err)
	}

	return nil
}
