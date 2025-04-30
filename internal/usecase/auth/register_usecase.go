package auth

import (
	"context"
	"mime/multipart"
)

type RegisterInput struct {
	Name     string                `json:"name" binding:"required,max=255"`
	Email    string                `json:"email" binding:"required,email,customEmail,max=255"`
	Password string                `json:"password" binding:"required,customPassword,min=6,max=255"`
	Avatar   *multipart.FileHeader `json:"avatar" binding:"required,omitempty"`
}

type RegisterOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func (a *authUseCase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	return nil, nil
}
