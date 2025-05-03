package profile

import (
	"context"
	"gochat-backend/internal/repository"
)

type ProfileOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type ProfileUseCase interface {
	GetUserProfile(ctx context.Context, userID string) (*ProfileOutput, error)
	SearchUsersByName(ctx context.Context, input SearchUsersInput) (*SearchUsersOutput, error)
}

type profileUseCase struct {
	accountRepository repository.AccountRepository
}

func NewProfileUseCase(accountRepository repository.AccountRepository) ProfileUseCase {
	return &profileUseCase{
		accountRepository: accountRepository,
	}
}
