package main

import (
	"fmt"
	"gochat-backend/config"
	"gochat-backend/internal/infra/mysqlinfra"
	"log"
	"os"
	"strconv"

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

	// Sử dụng hàm ExecuteMigration từ package mysqlinfra
	if err := mysqlinfra.ExecuteMigration(database, action, version); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
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
	fmt.Println("\nAlternatively, you can set the following environment variables:")
	fmt.Println("  MIGRATION_ACTION=up|down|reset|force|goto")
	fmt.Println("  MIGRATION_VERSION=N (required for force and goto actions)")
}
