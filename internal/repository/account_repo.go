package repository

import (
	"context"
	"database/sql"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/infra/mysqlinfra"
	"time"

	"github.com/google/uuid"
)

type AccountRepository interface {
	CreateUser(ctx context.Context, account *domain.Account) error
	FindByEmail(ctx context.Context, email string) (*domain.Account, error)
	FindByID(ctx context.Context, id string) (*domain.Account, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdatePassword(ctx context.Context, id string, password string) error
	UpdateProfileInfo(ctx context.Context, account *domain.Account) error
}

type AccountRepo struct {
	database *mysqlinfra.Database
}

func NewUserRepo(db *mysqlinfra.Database) *AccountRepo {
	return &AccountRepo{database: db}
}

func (r *AccountRepo) CreateUser(ctx context.Context, account *domain.Account) error {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}

	query := `
        INSERT INTO users (id, name, email, password, avatar_url, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			query,
			account.ID,
			account.Name,
			account.Email,
			account.Password,
			account.AvatarURL,
			time.Now(),
			time.Now(),
		)

		return err
	})
}

func (r *AccountRepo) FindByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var account domain.Account
	query := `
        SELECT 
            id, name, email, password, avatar_url, created_at, updated_at 
        FROM users 
        WHERE email = ?
    `

	err := r.database.DB.QueryRowContext(ctx, query, email).Scan(
		&account.ID,
		&account.Name,
		&account.Email,
		&account.Password,
		&account.AvatarURL,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepo) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	var account domain.Account
	query := `
        SELECT 
            id, name, email, password, avatar_url, created_at, updated_at 
        FROM users 
        WHERE id = ?
    `

	err := r.database.DB.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.Name,
		&account.Email,
		&account.Password,
		&account.AvatarURL,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found
		}
		return nil, err // Some other error occurred
	}

	return &account, nil
}

func (r *AccountRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = ?`
	var count int
	err := r.database.DB.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AccountRepo) UpdatePassword(ctx context.Context, id string, password string) error {
	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		query := `UPDATE users SET password = ?, updated_at = ? WHERE id = ?`
		_, err := tx.ExecContext(ctx, query, password, time.Now(), id)
		return err
	})
}

func (r *AccountRepo) UpdateProfileInfo(ctx context.Context, account *domain.Account) error {
	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		query := `
            UPDATE users SET 
                name = ?, 
                avatar_url = ?, 
                updated_at = ? 
            WHERE id = ?
        `
		_, err := tx.ExecContext(
			ctx,
			query,
			account.Name,
			account.AvatarURL,
			time.Now(),
			account.ID,
		)
		return err
	})
}
