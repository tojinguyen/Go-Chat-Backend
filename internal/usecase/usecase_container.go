package usecase

import (
	"gochat-backend/internal/config"
	"gochat-backend/internal/repository"
	"gochat-backend/internal/usecase/auth"
	"gochat-backend/pkg/jwt"
)

type SharedDependencies struct {
	Config      *config.Environment
	JwtService  jwt.JwtService
	AccountRepo repository.AccountRepository
}

type UseCaseContainer struct {
	Auth auth.AuthUseCase
}

func NewUseCaseContainer(deps *SharedDependencies) *UseCaseContainer {
	return &UseCaseContainer{
		Auth: auth.NewAuthUseCase(deps.Config, deps.JwtService, deps.AccountRepo),
	}
}
