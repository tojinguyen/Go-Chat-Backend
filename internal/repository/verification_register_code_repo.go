package repository

import (
	"context"
	"database/sql"
	"errors"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/infra/mysqlinfra"
	"time"

	"github.com/google/uuid"
)

type VerificationRegisterCodeRepository interface {
	CreateVerificationCode(ctx context.Context, code *domain.RegistrationVerificationCode) error
	GetVerificationCodeByID(ctx context.Context, id string) (*domain.RegistrationVerificationCode, error)
	GetVerificationCodeByEmail(ctx context.Context, email string) (*domain.RegistrationVerificationCode, error)
	VerifyCode(ctx context.Context, id string, code string) error
	UpdateVerificationStatus(ctx context.Context, id string, verified bool) error
	DeleteVerificationCode(ctx context.Context, id string) error
}

type registerVerificationRepo struct {
	database *mysqlinfra.Database
}

func NewVerificationRepo(db *mysqlinfra.Database) *registerVerificationRepo {
	return &registerVerificationRepo{database: db}
}

func (r *registerVerificationRepo) CreateVerificationCode(ctx context.Context, code *domain.RegistrationVerificationCode) error {
	if code.ID == "" {
		code.ID = uuid.New().String()
	}

	if code.CreatedAt.IsZero() {
		code.CreatedAt = time.Now()
	}

	query := `
        INSERT INTO verification_codes (id, user_id, email, name, hashed_password, avatar, code, type, verified, expires_at, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			query,
			code.ID,
			code.UserID,
			code.Email,
			code.Name,
			code.HashedPassword,
			code.Avatar,
			code.Code,
			code.Type,
			code.Verified,
			code.ExpiresAt,
			code.CreatedAt,
		)

		return err
	})
}

func (r *registerVerificationRepo) GetVerificationCodeByID(ctx context.Context, id string) (*domain.RegistrationVerificationCode, error) {
	query := `
        SELECT id, user_id, email, name, hashed_password, avatar, code, type, verified, expires_at, created_at 
        FROM verification_codes 
        WHERE id = ? AND type = 'register'
    `

	var code domain.RegistrationVerificationCode

	err := r.database.DB.QueryRowContext(ctx, query, id).Scan(
		&code.ID,
		&code.UserID,
		&code.Email,
		&code.Name,
		&code.HashedPassword,
		&code.Avatar,
		&code.Code,
		&code.Type,
		&code.Verified,
		&code.ExpiresAt,
		&code.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &code, nil
}

func (r *registerVerificationRepo) GetVerificationCodeByEmail(ctx context.Context, email string) (*domain.RegistrationVerificationCode, error) {
	query := `
        SELECT id, user_id, email, name, hashed_password, avatar, code, type, verified, expires_at, created_at 
        FROM verification_codes 
        WHERE email = ? AND type = 'register'
        ORDER BY created_at DESC LIMIT 1
    `

	var code domain.RegistrationVerificationCode

	err := r.database.DB.QueryRowContext(ctx, query, email).Scan(
		&code.ID,
		&code.UserID,
		&code.Email,
		&code.Name,
		&code.HashedPassword,
		&code.Avatar,
		&code.Code,
		&code.Type,
		&code.Verified,
		&code.ExpiresAt,
		&code.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &code, nil
}

func (r *registerVerificationRepo) VerifyCode(ctx context.Context, id string, code string) error {
	query := `
        SELECT id FROM verification_codes 
        WHERE id = ? AND code = ? AND type = 'register' AND verified = false AND expires_at > NOW()
    `

	var verificationID string
	err := r.database.DB.QueryRowContext(ctx, query, id, code).Scan(&verificationID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid or expired verification code")
		}
		return err
	}

	return r.UpdateVerificationStatus(ctx, id, true)
}

func (r *registerVerificationRepo) UpdateVerificationStatus(ctx context.Context, id string, verified bool) error {
	query := `UPDATE verification_codes SET verified = ? WHERE id = ?`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query, verified, id)
		return err
	})
}

func (r *registerVerificationRepo) DeleteVerificationCode(ctx context.Context, id string) error {
	query := `DELETE FROM verification_codes WHERE id = ?`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query, id)
		return err
	})
}
