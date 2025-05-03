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
	UpdateAvatar(ctx context.Context, id string, avatarURL string) error
	FindByName(ctx context.Context, name string, limit, offset int) ([]*domain.Account, error)
	CountByName(ctx context.Context, name string) (int, error)
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

func (r *AccountRepo) UpdateAvatar(ctx context.Context, id string, avatarURL string) error {
	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		query := `UPDATE users SET avatar_url = ?, updated_at = ? WHERE id = ?`
		_, err := tx.ExecContext(ctx, query, avatarURL, time.Now(), id)
		return err
	})
}

func (r *AccountRepo) FindByName(ctx context.Context, name string, limit, offset int) ([]*domain.Account, error) {
	query := `
        SELECT 
            id, name, email, password, avatar_url, created_at, updated_at 
        FROM users 
        WHERE name LIKE ? 
        LIMIT ? OFFSET ?
    `

	searchPattern := "%" + name + "%"
	rows, err := r.database.DB.QueryContext(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		var account domain.Account
		err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Email,
			&account.Password,
			&account.AvatarURL,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *AccountRepo) CountByName(ctx context.Context, name string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE name LIKE ?`

	searchPattern := "%" + name + "%"
	var count int
	err := r.database.DB.QueryRowContext(ctx, query, searchPattern).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
