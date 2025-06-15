package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/repository/mocks"
	"gochat-backend/pkg/jwt"
	jwtMocks "gochat-backend/pkg/jwt/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthUseCase_VerifyToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		inputToken     string
		setupMocks     func(*jwtMocks.MockJwtService, *mocks.MockAccountRepository)
		expectedResult *LoginOutput
		expectedError  string
	}{
		{
			name:       "successful_token_verification",
			inputToken: "Bearer valid-access-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateAccessToken", ctx, "valid-access-token").Return(claims, nil)

				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindById", ctx, "user-123").Return(account, nil)
			},
			expectedResult: &LoginOutput{
				AccessToken:  "valid-access-token",
				RefreshToken: "",
				UserId:       "user-123",
				Email:        "test@example.com",
				FullName:     "Test User",
				Role:         "user",
				AvatarUrl:    "https://example.com/avatar.jpg",
			},
			expectedError: "",
		},
		{
			name:       "token_without_bearer_prefix",
			inputToken: "valid-access-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateAccessToken", ctx, "valid-access-token").Return(claims, nil)

				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindById", ctx, "user-123").Return(account, nil)
			},
			expectedResult: &LoginOutput{
				AccessToken:  "valid-access-token",
				RefreshToken: "",
				UserId:       "user-123",
				Email:        "test@example.com",
				FullName:     "Test User",
				Role:         "user",
				AvatarUrl:    "https://example.com/avatar.jpg",
			},
			expectedError: "",
		},
		{
			name:       "invalid_token",
			inputToken: "Bearer invalid-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				jwtSvc.On("ValidateAccessToken", ctx, "invalid-token").Return(nil, errors.New("token is invalid"))
			},
			expectedResult: nil,
			expectedError:  "token không hợp lệ",
		},
		{
			name:       "user_not_found_in_database",
			inputToken: "Bearer valid-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateAccessToken", ctx, "valid-token").Return(claims, nil)
				accountRepo.On("FindById", ctx, "user-123").Return(nil, errors.New("user not found"))
			},
			expectedResult: nil,
			expectedError:  "không tìm thấy người dùng",
		},
		{
			name:       "user_account_is_nil",
			inputToken: "Bearer valid-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateAccessToken", ctx, "valid-token").Return(claims, nil)
				accountRepo.On("FindById", ctx, "user-123").Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  "không tìm thấy người dùng",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockJwtService := new(jwtMocks.MockJwtService)
			mockAccountRepo := new(mocks.MockAccountRepository)

			// Setup mock expectations
			tt.setupMocks(mockJwtService, mockAccountRepo)

			// Create use case with mocks
			authUseCase := &authUseCase{
				jwtService:        mockJwtService,
				accountRepository: mockAccountRepo,
			}

			// Execute
			result, err := authUseCase.VerifyToken(ctx, tt.inputToken)

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
		})
	}
}

func TestAuthUseCase_RefreshToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		inputToken     string
		setupMocks     func(*jwtMocks.MockJwtService, *mocks.MockAccountRepository)
		expectedResult *LoginOutput
		expectedError  string
	}{
		{
			name:       "successful_token_refresh",
			inputToken: "valid-refresh-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateRefreshToken", "valid-refresh-token").Return(claims, nil)

				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindById", ctx, "user-123").Return(account, nil)

				jwtInput := &jwt.GenerateTokenInput{
					UserId: "user-123",
					Email:  "test@example.com",
					Role:   "user",
				}

				jwtSvc.On("GenerateAccessToken", jwtInput).Return("new-access-token", nil)
				jwtSvc.On("GenerateRefreshToken", jwtInput).Return("new-refresh-token", nil)
			},
			expectedResult: &LoginOutput{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				UserId:       "user-123",
				Email:        "test@example.com",
				FullName:     "Test User",
				Role:         "user",
				AvatarUrl:    "https://example.com/avatar.jpg",
			},
			expectedError: "",
		},
		{
			name:       "invalid_refresh_token",
			inputToken: "invalid-refresh-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				jwtSvc.On("ValidateRefreshToken", "invalid-refresh-token").Return(nil, errors.New("token is invalid"))
			},
			expectedResult: nil,
			expectedError:  "refresh token không hợp lệ",
		},
		{
			name:       "user_not_found_during_refresh",
			inputToken: "valid-refresh-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateRefreshToken", "valid-refresh-token").Return(claims, nil)
				accountRepo.On("FindById", ctx, "user-123").Return(nil, errors.New("user not found"))
			},
			expectedResult: nil,
			expectedError:  "không tìm thấy người dùng",
		},
		{
			name:       "user_account_is_nil_during_refresh",
			inputToken: "valid-refresh-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateRefreshToken", "valid-refresh-token").Return(claims, nil)
				accountRepo.On("FindById", ctx, "user-123").Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  "người dùng không tồn tại",
		},
		{
			name:       "failed_to_generate_access_token",
			inputToken: "valid-refresh-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateRefreshToken", "valid-refresh-token").Return(claims, nil)

				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindById", ctx, "user-123").Return(account, nil)

				jwtInput := &jwt.GenerateTokenInput{
					UserId: "user-123",
					Email:  "test@example.com",
					Role:   "user",
				}

				jwtSvc.On("GenerateAccessToken", jwtInput).Return("", errors.New("failed to generate access token"))
			},
			expectedResult: nil,
			expectedError:  "không thể tạo access token",
		},
		{
			name:       "failed_to_generate_refresh_token",
			inputToken: "valid-refresh-token",
			setupMocks: func(jwtSvc *jwtMocks.MockJwtService, accountRepo *mocks.MockAccountRepository) {
				claims := &jwt.CustomJwtClaims{
					GenerateTokenInput: jwt.GenerateTokenInput{
						UserId: "user-123",
						Email:  "test@example.com",
						Role:   "user",
					},
				}

				jwtSvc.On("ValidateRefreshToken", "valid-refresh-token").Return(claims, nil)

				account := &domain.Account{
					Id:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					AvatarURL: "https://example.com/avatar.jpg",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				accountRepo.On("FindById", ctx, "user-123").Return(account, nil)

				jwtInput := &jwt.GenerateTokenInput{
					UserId: "user-123",
					Email:  "test@example.com",
					Role:   "user",
				}

				jwtSvc.On("GenerateAccessToken", jwtInput).Return("new-access-token", nil)
				jwtSvc.On("GenerateRefreshToken", jwtInput).Return("", errors.New("failed to generate refresh token"))
			},
			expectedResult: nil,
			expectedError:  "không thể tạo refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockJwtService := new(jwtMocks.MockJwtService)
			mockAccountRepo := new(mocks.MockAccountRepository)

			// Setup mock expectations
			tt.setupMocks(mockJwtService, mockAccountRepo)

			// Create use case with mocks
			authUseCase := &authUseCase{
				jwtService:        mockJwtService,
				accountRepository: mockAccountRepo,
			}

			// Execute
			result, err := authUseCase.RefreshToken(ctx, tt.inputToken)

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
		})
	}
}
