//go:build unit
// +build unit

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"gochat-backend/config"
	domain "gochat-backend/internal/domain/auth"
	redisMocks "gochat-backend/internal/infra/redisinfra/mocks"
	"gochat-backend/internal/repository/mocks"
	"gochat-backend/pkg/jwt"
	jwtMocks "gochat-backend/pkg/jwt/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()

	// Create a hashed password for testing
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	tests := []struct {
		name           string
		input          LoginInput
		setupMocks     func(*jwtMocks.MockJwtService, *mocks.MockAccountRepository, *redisMocks.MockRedisService)
		expectedResult *LoginOutput
		expectedError  string
	}{{
		name: "successful_login",
		input: LoginInput{
			Email:    "test@example.com",
			Password: password,
		},
		setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository, redisSvc *redisMocks.MockRedisService) {
			account := &domain.Account{
				Id:        "user-123",
				Name:      "Test User",
				Email:     "test@example.com",
				Password:  string(hashedPassword),
				AvatarURL: "https://example.com/avatar.jpg",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			accountRepo.On("FindByEmail", ctx, "test@example.com").Return(account, nil)

			jwtInput := &jwt.GenerateTokenInput{
				UserId: "user-123",
				Email:  "test@example.com",
				Role:   config.USER,
			}

			jwtSvc.On("GenerateAccessToken", jwtInput).Return("access-token-123", nil)
			jwtSvc.On("GenerateRefreshToken", jwtInput).Return("refresh-token-123", nil)

			// Mock Redis Set call
			redisSvc.On("Set", ctx, "refresh_token:user-123", "refresh-token-123", mock.AnythingOfType("time.Duration")).Return(nil)
		},
		expectedResult: &LoginOutput{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-123",
			UserId:       "user-123",
			Email:        "test@example.com",
			FullName:     "Test User",
			Role:         config.USER,
			AvatarUrl:    "https://example.com/avatar.jpg",
		},
		expectedError: "",
	},
		{
			name: "user_not_found",
			input: LoginInput{
				Email:    "nonexistent@example.com",
				Password: password,
			},
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository, redisSvc *redisMocks.MockRedisService) {
				accountRepo.On("FindByEmail", ctx, "nonexistent@example.com").Return(nil, errors.New("user not found"))
			},
			expectedResult: nil,
			expectedError:  "failed to find user",
		},
		{
			name: "user_account_is_nil",
			input: LoginInput{
				Email:    "test@example.com",
				Password: password,
			},
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository, redisSvc *redisMocks.MockRedisService) {
				accountRepo.On("FindByEmail", ctx, "test@example.com").Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  "invalid email or password",
		},
		{
			name: "invalid_password",
			input: LoginInput{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository, redisSvc *redisMocks.MockRedisService) {
				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindByEmail", ctx, "test@example.com").Return(account, nil)
			},
			expectedResult: nil,
			expectedError:  "invalid email or password",
		},
		{
			name: "failed_to_generate_access_token",
			input: LoginInput{
				Email:    "test@example.com",
				Password: password,
			},
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository, redisSvc *redisMocks.MockRedisService) {
				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindByEmail", ctx, "test@example.com").Return(account, nil)

				jwtInput := &jwt.GenerateTokenInput{
					UserId: "user-123",
					Email:  "test@example.com",
					Role:   config.USER,
				}

				jwtSvc.On("GenerateAccessToken", jwtInput).Return("", errors.New("failed to generate token"))
			},
			expectedResult: nil,
			expectedError:  "failed to generate access token",
		},
		{
			name: "failed_to_generate_refresh_token",
			input: LoginInput{
				Email:    "test@example.com",
				Password: password,
			},
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository, redisSvc *redisMocks.MockRedisService) {
				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindByEmail", ctx, "test@example.com").Return(account, nil)

				jwtInput := &jwt.GenerateTokenInput{
					UserId: "user-123",
					Email:  "test@example.com",
					Role:   config.USER,
				}

				jwtSvc.On("GenerateAccessToken", jwtInput).Return("access-token-123", nil)
				jwtSvc.On("GenerateRefreshToken", jwtInput).Return("", errors.New("failed to generate refresh token"))
			},
			expectedResult: nil,
			expectedError:  "failed to generate refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockJwtService := new(jwtMocks.MockJwtService)
			mockAccountRepo := new(mocks.MockAccountRepository)
			mockRedisService := new(redisMocks.MockRedisService)

			// Setup mock expectations
			tt.setupMocks(mockJwtService, mockAccountRepo, mockRedisService)
			// Create use case with mocks
			authUseCase := &authUseCase{
				cfg: &config.Environment{
					RefreshTokenExpireMinutes: 60,
				},
				jwtService:        mockJwtService,
				accountRepository: mockAccountRepo,
				redisService:      mockRedisService,
			}

			// Execute
			result, err := authUseCase.Login(ctx, tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
			// Verify all mock expectations were met
			mockJwtService.AssertExpectations(t)
			mockAccountRepo.AssertExpectations(t)
			mockRedisService.AssertExpectations(t)
		})
	}
}
