package auth

import (
	"context"
	"errors"
	"fmt"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/pkg/verification"
	"mime/multipart"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	exists, err := a.accountRepository.ExistsByEmail(ctx, input.Email)

	if err != nil {
		return nil, fmt.Errorf("failed to check if email exists: %w", err)
	}

	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var avatarURL string

	// if input.Avatar != nil {
	// 	avatarURL, err = a.fileService.UploadFile(ctx, input.Avatar, "avatars")
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	// 	}
	// }

	verificationCode := a.verificationService.GenerateCode()

	account := &domain.Account{
		ID:        uuid.New().String(),
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashedPassword),
		AvatarURL: avatarURL,
	}

	if err := a.accountRepository.CreateUser(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	if err := a.emailService.SendVerificationCode(input.Email, verificationCode, verification.VerificationCodeTypeRegister); err != nil {
		return nil, fmt.Errorf("failed to send verification code: %w", err)
	}

	return &RegisterOutput{
		ID:        account.ID,
		Name:      account.Name,
		Email:     account.Email,
		AvatarURL: account.AvatarURL,
	}, nil
}
