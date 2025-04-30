package repository

import (
	"context"
	"database/sql"
	domain "gochat-backend/internal/domain/auth"
)

type AccountRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *AccountRepo {
	return &AccountRepo{DB: db}
}

func (r *AccountRepo) CreateUser(ctx context.Context, account *domain.Account) error {
	query := `INSERT INTO users (id, email, password) VALUES (UUID(), ?, ?)`
	_, err := r.DB.ExecContext(ctx, query, account.ID, account.Email, account.Password)
	return err
}

func (r *AccountRepo) FindByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var account domain.Account
	query := `SELECT id, email, password FROM users WHERE email = ?`
	err := r.DB.QueryRowContext(ctx, query, email).Scan(&account.ID, &account.Email, &account.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found
		}
		return nil, err // Some other error occurred
	}

	return &account, nil
}
