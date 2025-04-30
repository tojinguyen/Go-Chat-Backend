package auth

import "context"

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
}

type authUseCase struct {
}

func NewAuthUseCase() AuthUseCase {
	return &authUseCase{}
}
