package main

import (
	"fmt"
	"gochat-backend/internal/config"
	"gochat-backend/internal/infra/mysqlinfra"
	"log"
	"os"
	"strconv"

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
	version := -1 // -1 là mặc định, không chỉ định version

	if len(os.Args) > 1 {
		action = os.Args[1]

		// Nếu có tham số thứ hai (số phiên bản), đọc nó
		if len(os.Args) > 2 && (action == "force" || action == "goto") {
			v, err := strconv.Atoi(os.Args[2])
			if err == nil {
				version = v
			} else {
				log.Fatalf("Invalid version number: %v", err)
			}
		}
	}

	// Hiển thị hướng dẫn nếu được yêu cầu
	if action == "help" {
		printHelp()
		return
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
	if err := runMigration(database, action, version); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migrations completed successfully")
}

func printHelp() {
	fmt.Println("Migration tool usage:")
	fmt.Println("go run cmd/migrate/main.go [command] [version]")
	fmt.Println("\nCommands:")
	fmt.Println("  up      - Apply all migrations")
	fmt.Println("  down    - Rollback all migrations")
	fmt.Println("  reset   - Rollback all migrations and apply them again")
	fmt.Println("  force N - Force set database version to N without running migrations")
	fmt.Println("  goto N  - Migrate to specific version N")
	fmt.Println("  help    - Show this help")
	fmt.Println("\nExamples:")
	fmt.Println("  go run cmd/migrate/main.go up")
	fmt.Println("  go run cmd/migrate/main.go down")
	fmt.Println("  go run cmd/migrate/main.go force 0")
	fmt.Println("  go run cmd/migrate/main.go goto 1")
}

func runMigration(db *mysqlinfra.Database, action string, version int) error {
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

	var migrationErr error

	switch action {
	case "up":
		migrationErr = m.Up()
		if migrationErr != nil {
			if migrationErr == migrate.ErrNoChange {
				fmt.Println("Database is already up to date")
				return nil
			}
			return fmt.Errorf("failed to apply migrations: %w", migrationErr)
		}
		fmt.Println("Successfully applied all migrations")

	case "down":
		migrationErr = m.Down()
		if migrationErr != nil {
			if migrationErr == migrate.ErrNoChange {
				fmt.Println("No migrations to rollback")
				return nil
			}
			return fmt.Errorf("failed to rollback migrations: %w", migrationErr)
		}
		fmt.Println("Successfully rolled back all migrations")

	case "reset":
		migrationErr = m.Down()
		if migrationErr != nil && migrationErr != migrate.ErrNoChange {
			return fmt.Errorf("failed to rollback migrations during reset: %w", migrationErr)
		}

		migrationErr = m.Up()
		if migrationErr != nil && migrationErr != migrate.ErrNoChange {
			return fmt.Errorf("failed to apply migrations during reset: %w", migrationErr)
		}
		fmt.Println("Successfully reset and applied all migrations")

	case "force":
		if version < 0 {
			return fmt.Errorf("force requires a valid version number")
		}
		migrationErr = m.Force(version)
		if migrationErr != nil {
			return fmt.Errorf("failed to force version to %d: %w", version, migrationErr)
		}
		fmt.Printf("Successfully forced version to %d\n", version)

	case "goto":
		if version < 0 {
			return fmt.Errorf("goto requires a valid version number")
		}
		migrationErr = m.Migrate(uint(version))
		if migrationErr != nil {
			if migrationErr == migrate.ErrNoChange {
				fmt.Printf("Database is already at version %d\n", version)
				return nil
			}
			return fmt.Errorf("failed to migrate to version %d: %w", version, migrationErr)
		}
		fmt.Printf("Successfully migrated to version %d\n", version)

	default:
		return fmt.Errorf("unknown action: %s (try 'help' for usage)", action)
	}

	return nil
}
