package usecase

import "gochat-backend/internal/usecase/auth"

type UseCaseContainer struct {
	Auth auth.AuthUseCase
}

func NewUseCaseContainer() *UseCaseContainer {
	return &UseCaseContainer{
		Auth: auth.NewAuthUseCase(),
	}
}
