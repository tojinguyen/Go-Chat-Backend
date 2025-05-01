package repository

import (
	"context"
	"database/sql"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/infra/mysqlinfra"
	"time"

	"github.com/google/uuid"
)

type VerificationRegisterCodeRepository interface {
	CreateVerificationCode(ctx context.Context, code *domain.RegistrationVerificationCode) error
	FindByEmailAndType(ctx context.Context, email string, codeType string) (*domain.RegistrationVerificationCode, error)
	FindByCodeAndType(ctx context.Context, code string, codeType string) (*domain.RegistrationVerificationCode, error)
	MarkAsVerified(ctx context.Context, id string) error
}

type VerificationRepo struct {
	database *mysqlinfra.Database
}

func NewVerificationRepo(db *mysqlinfra.Database) *VerificationRepo {
	return &VerificationRepo{database: db}
}

func (r *VerificationRepo) CreateVerificationCode(ctx context.Context, code *domain.RegistrationVerificationCode) error {
	if code.ID == "" {
		code.ID = uuid.New().String()
	}

	if code.CreatedAt.IsZero() {
		code.CreatedAt = time.Now()
	}

	query := `
        INSERT INTO verification_codes (id, user_id, email, code, type, verified, expires_at, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			query,
			code.ID,
			code.UserID,
			code.Email,
			code.Code,
			code.Type,
			code.Verified,
			code.ExpiresAt,
			code.CreatedAt,
		)

		return err
	})
}

func (r *VerificationRepo) FindByEmailAndType(ctx context.Context, email string, codeType string) (*domain.RegistrationVerificationCode, error) {
	var code domain.RegistrationVerificationCode
	query := `
        SELECT 
            id, user_id, email, code, type, verified, expires_at, created_at, verified_at
        FROM verification_codes 
        WHERE email = ? AND type = ?
        ORDER BY created_at DESC
        LIMIT 1
    `

	var verifiedAt sql.NullTime
	err := r.database.DB.QueryRowContext(ctx, query, email, codeType).Scan(
		&code.ID,
		&code.UserID,
		&code.Email,
		&code.Code,
		&code.Type,
		&code.Verified,
		&code.ExpiresAt,
		&code.CreatedAt,
		&verifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if verifiedAt.Valid {
		code.VerifiedAt = &verifiedAt.Time
	}

	return &code, nil
}

func (r *VerificationRepo) FindByCodeAndType(ctx context.Context, code string, codeType string) (*domain.RegistrationVerificationCode, error) {
	var verificationCode domain.RegistrationVerificationCode
	query := `
        SELECT 
            id, user_id, email, code, type, verified, expires_at, created_at, verified_at
        FROM verification_codes 
        WHERE code = ? AND type = ? AND verified = false AND expires_at > NOW()
        LIMIT 1
    `

	var verifiedAt sql.NullTime
	err := r.database.DB.QueryRowContext(ctx, query, code, codeType).Scan(
		&verificationCode.ID,
		&verificationCode.UserID,
		&verificationCode.Email,
		&verificationCode.Code,
		&verificationCode.Type,
		&verificationCode.Verified,
		&verificationCode.ExpiresAt,
		&verificationCode.CreatedAt,
		&verifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if verifiedAt.Valid {
		verificationCode.VerifiedAt = &verifiedAt.Time
	}

	return &verificationCode, nil
}

func (r *VerificationRepo) MarkAsVerified(ctx context.Context, id string) error {
	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		query := `UPDATE verification_codes SET verified = true, verified_at = NOW() WHERE id = ?`
		_, err := tx.ExecContext(ctx, query, id)
		return err
	})
}
