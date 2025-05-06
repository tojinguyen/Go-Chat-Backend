package profile

import (
	"context"
	"errors"
	"fmt"
)

func (p *profileUseCase) GetUserProfile(ctx context.Context, userID string) (*ProfileOutput, error) {
	account, err := p.accountRepository.FindById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if account == nil {
		return nil, errors.New("user not found")
	}

	return &ProfileOutput{
		ID:        account.ID,
		Name:      account.Name,
		Email:     account.Email,
		AvatarURL: account.AvatarURL,
	}, nil
}
