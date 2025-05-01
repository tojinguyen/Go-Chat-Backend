package auth

import (
	"context"
	"gochat-backend/internal/config"
	"gochat-backend/internal/repository"
	"gochat-backend/pkg/email"
	"gochat-backend/pkg/jwt"
	"gochat-backend/pkg/verification"
)

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
}

type authUseCase struct {
	cfg                 *config.Environment
	jwtService          jwt.JwtService
	emailService        email.EmailService
	verificationService verification.VerificationService
	accountRepository   repository.AccountRepository
}

func NewAuthUseCase(
	cfg *config.Environment,
	jwtService jwt.JwtService,
	emailService email.EmailService,
	verificationService verification.VerificationService,
	accountRepository repository.AccountRepository,
) AuthUseCase {
	return &authUseCase{
		cfg:                 cfg,
		jwtService:          jwtService,
		emailService:        emailService,
		verificationService: verificationService,
		accountRepository:   accountRepository,
	}
}
