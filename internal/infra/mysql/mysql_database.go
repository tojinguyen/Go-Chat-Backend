package mysql

import (
	"database/sql"
	"fmt"
	"gochat-backend/internal/config"
	"time"
)

type Database struct {
	DB *sql.DB
}

func NewMySqlDatabase(db *sql.DB) *Database {
	return &Database{
		DB: db,
	}
}

func ConnectMysql(cfg *config.Environment) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local",
		cfg.MysqlUser,
		cfg.MysqlPassword,
		cfg.MysqlHost,
		cfg.MysqlPort,
		cfg.MysqlDatabase,
	)

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	// Thiết lập các tham số cho connection pool
	db.SetMaxOpenConns(cfg.MysqlMaxOpenConns)
	db.SetMaxIdleConns(cfg.MysqlMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.MysqlConnMaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.MysqlConnMaxIdleTime) * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	return db, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

func (d *Database) ExecuteTransaction(txFunc func(*sql.Tx) error) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		}
	}()

	if err := txFunc(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
