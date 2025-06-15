//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"gochat-backend/config"
	"gochat-backend/internal/infra/mysqlinfra"
	"gochat-backend/internal/infra/redisinfra"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

var (
	TestEnv      *config.Environment
	TestDB       *sql.DB
	TestRedis    *redis.Client
	MySQLService *mysqlinfra.Database
	RedisService redisinfra.RedisService
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	var err error

	// Load test environment
	TestEnv, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load test environment: %v", err)
	}

	// Setup test database
	if err := setupTestDatabase(); err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}

	// Setup test Redis
	if err := setupTestRedis(); err != nil {
		log.Fatalf("Failed to setup test Redis: %v", err)
	}

	log.Println("Integration test setup completed successfully")
}

func teardown() {
	if TestDB != nil {
		// Clean up test data
		cleanupTestData()
		TestDB.Close()
	}

	if TestRedis != nil {
		TestRedis.Close()
	}

	log.Println("Integration test teardown completed")
}

func setupTestDatabase() error {
	// Create database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		TestEnv.MysqlUser,
		TestEnv.MysqlPassword,
		TestEnv.MysqlHost,
		TestEnv.MysqlPort,
		TestEnv.MysqlDatabase,
	)

	var err error
	TestDB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %v", err)
	}

	// Test connection
	if err := TestDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping test database: %v", err)
	}

	// Configure connection pool
	TestDB.SetMaxOpenConns(TestEnv.MysqlMaxOpenConns)
	TestDB.SetMaxIdleConns(TestEnv.MysqlMaxIdleConns)
	TestDB.SetConnMaxLifetime(time.Duration(TestEnv.MysqlConnMaxLifetime) * time.Second)
	TestDB.SetConnMaxIdleTime(time.Duration(TestEnv.MysqlConnMaxIdleTime) * time.Second)
	// Create MySQL service
	MySQLService = mysqlinfra.NewMySqlDatabase(TestDB)

	return nil
}

func setupTestRedis() error {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", TestEnv.RedisHost, TestEnv.RedisPort),
		Password: TestEnv.RedisPassword,
		DB:       TestEnv.RedisDB}) // Test connection
	ctx, cancel := timeoutContext()
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping test Redis: %v", err)
	}

	TestRedis = rdb

	// Create Redis service using the public constructor
	var err error
	RedisService, err = redisinfra.NewRedisService(TestEnv)
	if err != nil {
		return fmt.Errorf("failed to create Redis service: %v", err)
	}

	return nil
}

func cleanupTestData() {
	tables := []string{
		"verification_register_code",
		"friend_requests",
		"friend_ships",
		"message",
		"chat_rooms",
		"account",
		"migrations",
	}

	for _, table := range tables {
		_, err := TestDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			log.Printf("Warning: Failed to cleanup table %s: %v", table, err)
		}
	}
	// Clean up Redis
	if TestRedis != nil {
		ctx, cancel := timeoutContext()
		defer cancel()
		TestRedis.FlushDB(ctx)
	}
}

func timeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
