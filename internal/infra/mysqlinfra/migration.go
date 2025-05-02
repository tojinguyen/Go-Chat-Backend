package mysqlinfra

import (
	"fmt"
	"gochat-backend/internal/config"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(cfg *config.Environment, db *Database) error {
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/mysql", // path to migration files
		"mysql",                   // database name
		driver,                    // database driver
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Chạy migration lên phiên bản mới nhất
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	log.Println("Database migrations applied successfully")

	return nil
}
