package mysqlinfra

import (
	"fmt"
	"gochat-backend/config"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationAction định nghĩa các hành động có thể thực hiện với migrations
type MigrationAction string

const (
	Up    MigrationAction = "up"    // Áp dụng tất cả migrations
	Down  MigrationAction = "down"  // Rollback tất cả migrations
	Reset MigrationAction = "reset" // Rollback và áp dụng lại migrations
	Force MigrationAction = "force" // Đặt version migrations về một giá trị cụ thể
	Goto  MigrationAction = "goto"  // Di chuyển đến version migrations cụ thể
)

// RunMigrations thực hiện migrations với hành động mặc định "up" (áp dụng tất cả migrations)
func RunMigrations(cfg *config.Environment, db *Database) error {
	// Lấy mode từ biến môi trường hoặc mặc định là "up"
	action := MigrationAction(os.Getenv("MIGRATION_ACTION"))
	if action == "" {
		action = Up
	}

	// Lấy version từ biến môi trường nếu cần
	version := -1
	if action == Force || action == Goto {
		if os.Getenv("MIGRATION_VERSION") != "" {
			fmt.Sscanf(os.Getenv("MIGRATION_VERSION"), "%d", &version)
		}
		if version < 0 {
			return fmt.Errorf("%s action requires a valid version number set in MIGRATION_VERSION", action)
		}
	}

	return ExecuteMigration(db, string(action), version)
}

// ExecuteMigration thực hiện migration với hành động và version được chỉ định
func ExecuteMigration(db *Database, action string, version int) error {
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
		if err != nil {
			if err == migrate.ErrNoChange {
				log.Println("Database is already up to date")
				return nil
			}
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
		log.Println("Successfully applied all migrations")

	case "down":
		err = m.Down()
		if err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No migrations to rollback")
				return nil
			}
			return fmt.Errorf("failed to rollback migrations: %w", err)
		}
		log.Println("Successfully rolled back all migrations")

	case "reset":
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to rollback migrations during reset: %w", err)
		}

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to apply migrations during reset: %w", err)
		}
		log.Println("Successfully reset and applied all migrations")

	case "force":
		if version < 0 {
			return fmt.Errorf("force requires a valid version number")
		}
		err = m.Force(version)
		if err != nil {
			return fmt.Errorf("failed to force version to %d: %w", version, err)
		}
		log.Printf("Successfully forced version to %d", version)

	case "goto":
		if version < 0 {
			return fmt.Errorf("goto requires a valid version number")
		}
		err = m.Migrate(uint(version))
		if err != nil {
			if err == migrate.ErrNoChange {
				log.Printf("Database is already at version %d", version)
				return nil
			}
			return fmt.Errorf("failed to migrate to version %d: %w", version, err)
		}
		log.Printf("Successfully migrated to version %d", version)

	default:
		return fmt.Errorf("unknown action: %s (use up, down, reset, force, or goto)", action)
	}

	return nil
}
