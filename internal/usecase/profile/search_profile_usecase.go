package profile

import (
	"context"
	"fmt"
)

type SearchUsersInput struct {
	Name  string `json:"name" binding:"required,min=1,max=255"`
	Page  int    `json:"page" binding:"min=1"`
	Limit int    `json:"limit" binding:"min=1,max=100"`
}

type UserItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type SearchUsersOutput struct {
	Users      []UserItem `json:"users"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

func (p *profileUseCase) SearchUsersByName(ctx context.Context, input SearchUsersInput) (*SearchUsersOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 10
	}

	offset := (input.Page - 1) * input.Limit

	// Get total count
	totalCount, err := p.accountRepository.CountByName(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// If no users found, return empty result
	if totalCount == 0 {
		return &SearchUsersOutput{
			Users:      []UserItem{},
			TotalCount: 0,
			Page:       input.Page,
			Limit:      input.Limit,
		}, nil
	}

	// Get users
	users, err := p.accountRepository.FindByName(ctx, input.Name, input.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Map to output format
	var userItems []UserItem
	for _, user := range users {
		userItems = append(userItems, UserItem{
			ID:        user.Id,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		})
	}

	return &SearchUsersOutput{
		Users:      userItems,
		TotalCount: totalCount,
		Page:       input.Page,
		Limit:      input.Limit,
	}, nil
}
