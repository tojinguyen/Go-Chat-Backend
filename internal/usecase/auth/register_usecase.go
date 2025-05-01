package auth

import (
	"context"
	"errors"
	"fmt"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/pkg/verification"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string                `json:"name" binding:"required,max=255"`
	Email    string                `json:"email" binding:"required,email,customEmail,max=255"`
	Password string                `json:"password" binding:"required,customPassword,min=6,max=255"`
	Avatar   *multipart.FileHeader `json:"avatar" binding:"required,omitempty"`
}

type VerifyRegistrationInput struct {
	ID    string `json:"id" binding:"required"`
	Email string `json:"email" binding:"required,email,customEmail,max=255"`
	Code  string `json:"code" binding:"required"`
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

	if input.Avatar != nil {
		avatarURL, err = a.cloudstorage.UploadAvatar(input.Avatar, "avatars/temp")
		if err != nil {
			return nil, fmt.Errorf("failed to upload avatar: %w", err)
		}
	}

	userID := uuid.New().String()
	verificationCode := a.verificationService.GenerateCode()

	expiresAt := time.Now().UTC().Add(time.Duration(a.cfg.VerificationCodeExpireMinutes) * time.Minute)

	verificationRecord := &domain.RegistrationVerificationCode{
		ID:             uuid.New().String(),
		UserID:         userID,
		Email:          input.Email,
		Name:           input.Name,
		HashedPassword: string(hashedPassword),
		Avatar:         avatarURL,
		Code:           verificationCode,
		Type:           verification.VerificationCodeTypeRegister,
		Verified:       false,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
	}

	err = a.verificationRegisterRepository.CreateVerificationCode(ctx, verificationRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to save verification data: %w", err)
	}

	if err := a.emailService.SendVerificationCode(input.Email, verificationCode, verification.VerificationCodeTypeRegister); err != nil {
		return nil, fmt.Errorf("failed to send verification code: %w", err)
	}

	return &RegisterOutput{
		ID:        userID,
		Name:      input.Name,
		Email:     input.Email,
		AvatarURL: avatarURL,
	}, nil
}

func (a *authUseCase) VerifyRegistration(ctx context.Context, input VerifyRegistrationInput) (*RegisterOutput, error) {

	return nil, nil
}
