package repository

import (
	"context"
	"database/sql"
	"fmt"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/infra/mysqlinfra"
	"gochat-backend/internal/infra/redisinfra"
	"time"

	"github.com/google/uuid"
)

const (
	userCacheKeyPrefix = "user:"
	userCacheTTLExpiry = 24 * time.Hour
)

type AccountRepository interface {
	CreateUser(ctx context.Context, account *domain.Account) error
	FindByEmail(ctx context.Context, email string) (*domain.Account, error)
	FindById(ctx context.Context, id string) (*domain.Account, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdatePassword(ctx context.Context, id string, password string) error
	UpdateProfileInfo(ctx context.Context, account *domain.Account) error
	UpdateAvatar(ctx context.Context, id string, avatarURL string) error
	FindByName(ctx context.Context, name string, limit, offset int) ([]*domain.Account, error)
	CountByName(ctx context.Context, name string) (int, error)
}

type accountRepo struct {
	database     *mysqlinfra.Database
	redisService redisinfra.RedisService
}

func NewAccountRepo(db *mysqlinfra.Database, redisService redisinfra.RedisService) AccountRepository {
	return &accountRepo{
		database:     db,
		redisService: redisService,
	}
}

func (r *accountRepo) CreateUser(ctx context.Context, account *domain.Account) error {
	if account.Id == "" {
		account.Id = uuid.New().String()
	}

	query := `
        INSERT INTO users (id, name, email, password, avatar_url, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`

	err := r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			query,
			account.Id,
			account.Name,
			account.Email,
			account.Password,
			account.AvatarURL,
			time.Now(),
			time.Now(),
		)

		return err
	})

	if err != nil {
		return err
	}

	cacheKey := r.generateUserCacheKey(account.Id)
	accountToCache := *account
	accountToCache.Password = ""
	if err := r.redisService.Set(ctx, cacheKey, &accountToCache, userCacheTTLExpiry); err != nil {
		fmt.Printf("Warning: Failed to cache user after creation (ID: %s): %v\n", account.Id, err)
	}
	return nil
}

func (r *accountRepo) FindByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var account domain.Account
	query := `
        SELECT 
            id, name, email, password, avatar_url, created_at, updated_at 
        FROM users 
        WHERE email = ?
    `

	err := r.database.DB.QueryRowContext(ctx, query, email).Scan(
		&account.Id,
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

func (r *accountRepo) FindById(ctx context.Context, id string) (*domain.Account, error) {
	cacheKey := r.generateUserCacheKey(id)
	var cachedAccount domain.Account

	// 1. Try to get the account from cache
	if err := r.redisService.Get(ctx, cacheKey, &cachedAccount); err == nil {
		return &cachedAccount, nil
	}

	// 2. If not found in cache, query the database
	var account domain.Account
	query := `
        SELECT 
            id, name, email, password, avatar_url, created_at, updated_at 
        FROM users 
        WHERE id = ?
    `

	err := r.database.DB.QueryRowContext(ctx, query, id).Scan(
		&account.Id,
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

	// 3. Save the account to cache
	accountToCache := account
	accountToCache.Password = ""
	if err := r.redisService.Set(ctx, cacheKey, &accountToCache, userCacheTTLExpiry); err != nil {
		fmt.Printf("Warning: Failed to cache user (ID: %s): %v\n", id, err)
	}

	return &account, nil
}

func (r *accountRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = ?`
	var count int
	err := r.database.DB.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *accountRepo) UpdatePassword(ctx context.Context, id string, password string) error {
	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		query := `UPDATE users SET password = ?, updated_at = ? WHERE id = ?`
		_, err := tx.ExecContext(ctx, query, password, time.Now(), id)
		return err
	})
}

func (r *accountRepo) UpdateProfileInfo(ctx context.Context, account *domain.Account) error {
	err := r.database.ExecuteTransaction(func(tx *sql.Tx) error {
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
			account.Id,
		)
		return err
	})
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := r.generateUserCacheKey(account.Id)
	if err := r.redisService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Warning: Failed to delete user cache after update (ID: %s): %v\n", account.Id, err)
	}
	return nil
}

func (r *accountRepo) UpdateAvatar(ctx context.Context, id string, avatarURL string) error {
	err := r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		query := `UPDATE users SET avatar_url = ?, updated_at = ? WHERE id = ?`
		_, err := tx.ExecContext(ctx, query, avatarURL, time.Now(), id)
		return err
	})
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := r.generateUserCacheKey(id)
	if err := r.redisService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Warning: Failed to delete user cache after avatar update (ID: %s): %v\n", id, err)
	}
	return nil
}

func (r *accountRepo) FindByName(ctx context.Context, name string, limit, offset int) ([]*domain.Account, error) {
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
			&account.Id,
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

func (r *accountRepo) CountByName(ctx context.Context, name string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE name LIKE ?`

	searchPattern := "%" + name + "%"
	var count int
	err := r.database.DB.QueryRowContext(ctx, query, searchPattern).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *accountRepo) generateUserCacheKey(id string) string {
	return fmt.Sprintf("%s%s", userCacheKeyPrefix, id)
}
