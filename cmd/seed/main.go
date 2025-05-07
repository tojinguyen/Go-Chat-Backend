package main

import (
	"database/sql"
	"fmt"
	"gochat-backend/config"
	"gochat-backend/internal/infra/mysqlinfra"
	"log"
	"strings"
	"time"

	"github.com/bxcodec/faker/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID        string
	Name      string
	AvatarURL string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func generateFakeUser() Account {
	now := time.Now()
	password := faker.Password()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		// Fallback to plain password in case of error
		hashedPassword = []byte(password)
	}

	return Account{
		ID:        faker.UUIDDigit(),
		Name:      faker.Name(),
		AvatarURL: faker.URL(),
		Email:     faker.Email(),
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func insertBatch(db *sql.DB, accounts []Account) error {
	if len(accounts) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString("INSERT INTO accounts (id, name, avatar_url, email, password, created_at, updated_at) VALUES ")

	args := make([]interface{}, 0, len(accounts)*7)

	for i, acc := range accounts {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			acc.ID, acc.Name, acc.AvatarURL, acc.Email, acc.Password, acc.CreatedAt, acc.UpdatedAt,
		)
	}

	stmt := builder.String()
	_, err := db.Exec(stmt, args...)
	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
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

	totalUsers := 100000
	batchSize := 1000
	var accounts []Account

	start := time.Now()
	for i := 0; i < totalUsers; i++ {
		accounts = append(accounts, generateFakeUser())
		if len(accounts) >= batchSize {
			err := insertBatch(db, accounts)
			if err != nil {
				log.Fatalf("Batch insert failed at user %d: %v", i, err)
			}
			accounts = accounts[:0]
			fmt.Printf("Inserted %d users...\n", i+1)
		}
	}
	if len(accounts) > 0 {
		if err := insertBatch(db, accounts); err != nil {
			log.Fatalf("Final batch insert failed: %v", err)
		}
	}

	fmt.Println("✅ Done! Total time:", time.Since(start))
}
