package auth

import (
	"context"
	"errors"
	"fmt"
	"gochat-backend/config"
	"gochat-backend/pkg/jwt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required,email,customEmail,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

type LoginOutput struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserId       string `json:"userId"`
	Email        string `json:"email"`
	FullName     string `json:"fullName"`
	Role         string `json:"role"`
	AvatarUrl    string `json:"avatarUrl"`
}

func (a *authUseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// Find the user by email
	account, err := a.accountRepository.FindByEmail(ctx, input.Email)
	if err != nil {
		log.Printf("Error finding user by email: %v\n", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if account == nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare the passwords
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(input.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT tokens
	jwtInput := &jwt.GenerateTokenInput{
		UserId: account.Id, // Using string ID instead
		Email:  account.Email,
		Role:   config.USER, // Assuming default role is USER
	}

	accessToken, err := a.jwtService.GenerateAccessToken(jwtInput)
	if err != nil {
		log.Printf("Error generating access token: %v\n", err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := a.jwtService.GenerateRefreshToken(jwtInput)
	if err != nil {
		log.Printf("Error generating refresh token: %v\n", err)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis with user ID as key for later validation
	// Using a TTL that matches the refresh token expiration
	refreshTokenKey := fmt.Sprintf("refresh_token:%s", account.Id)
	err = a.redisService.Set(ctx, refreshTokenKey, refreshToken, time.Duration(a.cfg.RefreshTokenExpireMinutes)*time.Minute)
	if err != nil {
		log.Printf("Error storing refresh token in Redis: %v\n", err)
		// Continue anyway as this is not critical - token will still work but can't be invalidated
	}

	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       account.Id,
		Email:        account.Email,
		FullName:     account.Name,
		Role:         config.USER,
		AvatarUrl:    account.AvatarURL,
	}, nil
}
