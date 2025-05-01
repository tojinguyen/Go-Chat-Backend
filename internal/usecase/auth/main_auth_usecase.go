package auth

import (
	"context"
	"gochat-backend/internal/config"
	"gochat-backend/internal/repository"
	"gochat-backend/pkg/jwt"
)

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
}

type authUseCase struct {
	cfg               *config.Environment
	jwtService        jwt.JwtService
	accountRepository repository.AccountRepository
}

func NewAuthUseCase(
	cfg *config.Environment,
	jwtService jwt.JwtService,
	accountRepository repository.AccountRepository,
) AuthUseCase {
	return &authUseCase{
		cfg:               cfg,
		jwtService:        jwtService,
		accountRepository: accountRepository,
	}
}
