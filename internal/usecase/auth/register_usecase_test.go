package auth

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"gochat-backend/config"
	domain "gochat-backend/internal/domain/auth"
	cloudinaryMocks "gochat-backend/internal/infra/cloudinaryinfra/mocks"
	"gochat-backend/internal/repository/mocks"
	emailMocks "gochat-backend/pkg/email/mocks"
	"gochat-backend/pkg/verification"
	verificationMocks "gochat-backend/pkg/verification/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthUseCase_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		input          RegisterInput
		setupMocks     func(*mocks.MockAccountRepository, *mocks.MockVerificationRegisterCodeRepository, *verificationMocks.MockVerificationService, *emailMocks.MockEmailService, *cloudinaryMocks.MockCloudinaryService)
		expectedResult *RegisterOutput
		expectedError  string
	}{
		{
			name: "successful_registration_with_avatar",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
				Avatar:   &multipart.FileHeader{Filename: "avatar.jpg"},
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, verificationSvc *verificationMocks.MockVerificationService, emailSvc *emailMocks.MockEmailService, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Check email doesn't exist
				accountRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)

				// Check no existing verification record
				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(nil, nil)

				// Avatar upload
				cloudSvc.On("UploadAvatar", &multipart.FileHeader{Filename: "avatar.jpg"}, "avatars/temp").Return("https://cloudinary.com/avatar.jpg", nil)

				// Generate verification code
				verificationSvc.On("GenerateCode").Return("123456")

				// Save verification code
				verificationRepo.On("CreateVerificationCode", ctx, mock.MatchedBy(func(code *domain.RegistrationVerificationCode) bool {
					return code.Email == "test@example.com" && code.Name == "Test User" && code.Code == "123456"
				})).Return(nil)

				// Send email
				emailSvc.On("SendVerificationCode", "test@example.com", "123456", verification.VerificationCodeTypeRegister).Return(nil)
			},
			expectedResult: &RegisterOutput{
				Name:      "Test User",
				Email:     "test@example.com",
				AvatarURL: "https://cloudinary.com/avatar.jpg",
			},
			expectedError: "",
		},
		{
			name: "successful_registration_without_avatar",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
				Avatar:   nil,
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, verificationSvc *verificationMocks.MockVerificationService, emailSvc *emailMocks.MockEmailService, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Check email doesn't exist
				accountRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)

				// Check no existing verification record
				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(nil, nil)

				// Generate verification code
				verificationSvc.On("GenerateCode").Return("123456")

				// Save verification code
				verificationRepo.On("CreateVerificationCode", ctx, mock.MatchedBy(func(code *domain.RegistrationVerificationCode) bool {
					return code.Email == "test@example.com" && code.Name == "Test User" && code.Code == "123456"
				})).Return(nil)

				// Send email
				emailSvc.On("SendVerificationCode", "test@example.com", "123456", verification.VerificationCodeTypeRegister).Return(nil)
			},
			expectedResult: &RegisterOutput{
				Name:      "Test User",
				Email:     "test@example.com",
				AvatarURL: "https://default-avatar.com/avatar.jpg", // Default avatar from config
			},
			expectedError: "",
		},
		{
			name: "email_already_exists",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, verificationSvc *verificationMocks.MockVerificationService, emailSvc *emailMocks.MockEmailService, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				accountRepo.On("ExistsByEmail", ctx, "existing@example.com").Return(true, nil)
			},
			expectedResult: nil,
			expectedError:  "email already exists",
		},
		{
			name: "error_checking_email_exists",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, verificationSvc *verificationMocks.MockVerificationService, emailSvc *emailMocks.MockEmailService, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				accountRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, errors.New("database error"))
			},
			expectedResult: nil,
			expectedError:  "failed to check if email exists",
		},
		{
			name: "delete_existing_verification_record",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, verificationSvc *verificationMocks.MockVerificationService, emailSvc *emailMocks.MockEmailService, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Check email doesn't exist
				accountRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)

				// Existing verification record
				existingRecord := &domain.RegistrationVerificationCode{
					ID:    "existing-id",
					Email: "test@example.com",
				}
				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(existingRecord, nil)
				verificationRepo.On("DeleteVerificationCode", ctx, "existing-id").Return(nil)

				// Generate verification code
				verificationSvc.On("GenerateCode").Return("123456")

				// Save verification code
				verificationRepo.On("CreateVerificationCode", ctx, mock.MatchedBy(func(code *domain.RegistrationVerificationCode) bool {
					return code.Email == "test@example.com" && code.Name == "Test User" && code.Code == "123456"
				})).Return(nil)

				// Send email
				emailSvc.On("SendVerificationCode", "test@example.com", "123456", verification.VerificationCodeTypeRegister).Return(nil)
			},
			expectedResult: &RegisterOutput{
				Name:      "Test User",
				Email:     "test@example.com",
				AvatarURL: "https://default-avatar.com/avatar.jpg",
			},
			expectedError: "",
		},
		{
			name: "failed_to_upload_avatar",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
				Avatar:   &multipart.FileHeader{Filename: "avatar.jpg"},
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, verificationSvc *verificationMocks.MockVerificationService, emailSvc *emailMocks.MockEmailService, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Check email doesn't exist
				accountRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)

				// Check no existing verification record
				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(nil, nil)

				// Avatar upload fails
				cloudSvc.On("UploadAvatar", &multipart.FileHeader{Filename: "avatar.jpg"}, "avatars/temp").Return("", errors.New("upload failed"))
			},
			expectedResult: nil,
			expectedError:  "failed to upload avatar",
		},
		{
			name: "failed_to_send_email",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, verificationSvc *verificationMocks.MockVerificationService, emailSvc *emailMocks.MockEmailService, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Check email doesn't exist
				accountRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)

				// Check no existing verification record
				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(nil, nil)

				// Generate verification code
				verificationSvc.On("GenerateCode").Return("123456")

				// Save verification code
				verificationRepo.On("CreateVerificationCode", ctx, mock.MatchedBy(func(code *domain.RegistrationVerificationCode) bool {
					return code.Email == "test@example.com" && code.Name == "Test User" && code.Code == "123456"
				})).Return(nil)

				// Send email fails
				emailSvc.On("SendVerificationCode", "test@example.com", "123456", verification.VerificationCodeTypeRegister).Return(errors.New("email failed"))
			},
			expectedResult: nil,
			expectedError:  "failed to send verification code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAccountRepo := new(mocks.MockAccountRepository)
			mockVerificationRepo := new(mocks.MockVerificationRegisterCodeRepository)
			mockVerificationService := new(verificationMocks.MockVerificationService)
			mockEmailService := new(emailMocks.MockEmailService)
			mockCloudService := new(cloudinaryMocks.MockCloudinaryService)

			// Setup mock expectations
			tt.setupMocks(mockAccountRepo, mockVerificationRepo, mockVerificationService, mockEmailService, mockCloudService)
			// Create use case with mocks
			authUseCase := &authUseCase{
				cfg: &config.Environment{
					Constants: config.Constants{
						DefaultAvatarURL: "https://default-avatar.com/avatar.jpg",
					},
				},
				accountRepository:              mockAccountRepo,
				verificationRegisterRepository: mockVerificationRepo,
				verificationService:            mockVerificationService,
				emailService:                   mockEmailService,
				cloudstorage:                   mockCloudService,
			}

			// Execute
			result, err := authUseCase.Register(ctx, tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Name, result.Name)
				assert.Equal(t, tt.expectedResult.Email, result.Email)
				assert.Equal(t, tt.expectedResult.AvatarURL, result.AvatarURL)
				assert.NotEmpty(t, result.ID) // ID should be generated
			}

			// Verify all mock expectations were met
			mockAccountRepo.AssertExpectations(t)
			mockVerificationRepo.AssertExpectations(t)
			mockVerificationService.AssertExpectations(t)
			mockEmailService.AssertExpectations(t)
			mockCloudService.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_VerifyRegistration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		input          VerifyRegistrationInput
		setupMocks     func(*mocks.MockAccountRepository, *mocks.MockVerificationRegisterCodeRepository, *cloudinaryMocks.MockCloudinaryService)
		expectedResult *RegisterOutput
		expectedError  string
	}{
		{
			name: "successful_verification",
			input: VerifyRegistrationInput{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Verification record exists and is valid
				verificationRecord := &domain.RegistrationVerificationCode{
					ID:             "verification-id",
					Email:          "test@example.com",
					Name:           "Test User",
					HashedPassword: "hashed-password",
					Avatar:         "https://cloudinary.com/avatar.jpg",
					Code:           "123456",
					Type:           verification.VerificationCodeTypeRegister,
					ExpiresAt:      time.Now().Add(10 * time.Minute), // Valid for 10 more minutes
					CreatedAt:      time.Now(),
				}

				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(verificationRecord, nil)

				// Create user account
				accountRepo.On("CreateUser", ctx, mock.MatchedBy(func(account *domain.Account) bool {
					return account.Email == "test@example.com" && account.Name == "Test User"
				})).Return(nil)

				// Delete verification record
				verificationRepo.On("DeleteVerificationCode", ctx, "verification-id").Return(nil)
			},
			expectedResult: &RegisterOutput{
				Name:      "Test User",
				Email:     "test@example.com",
				AvatarURL: "https://cloudinary.com/avatar.jpg",
			},
			expectedError: "",
		},
		{
			name: "verification_record_not_found",
			input: VerifyRegistrationInput{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  "verification record not found",
		},
		{
			name: "verification_code_expired",
			input: VerifyRegistrationInput{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Expired verification record
				verificationRecord := &domain.RegistrationVerificationCode{
					ID:             "verification-id",
					Email:          "test@example.com",
					Name:           "Test User",
					HashedPassword: "hashed-password",
					Avatar:         "https://cloudinary.com/avatar.jpg",
					Code:           "123456",
					Type:           verification.VerificationCodeTypeRegister,
					ExpiresAt:      time.Now().Add(-10 * time.Minute), // Expired 10 minutes ago
					CreatedAt:      time.Now(),
				}

				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(verificationRecord, nil)
			},
			expectedResult: nil,
			expectedError:  "verification code has expired",
		},
		{
			name: "invalid_verification_code",
			input: VerifyRegistrationInput{
				Email: "test@example.com",
				Code:  "wrong-code",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Valid verification record but wrong code
				verificationRecord := &domain.RegistrationVerificationCode{
					ID:             "verification-id",
					Email:          "test@example.com",
					Name:           "Test User",
					HashedPassword: "hashed-password",
					Avatar:         "https://cloudinary.com/avatar.jpg",
					Code:           "123456", // Correct code
					Type:           verification.VerificationCodeTypeRegister,
					ExpiresAt:      time.Now().Add(10 * time.Minute),
					CreatedAt:      time.Now(),
				}

				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(verificationRecord, nil)
			},
			expectedResult: nil,
			expectedError:  "invalid verification code",
		},
		{
			name: "failed_to_create_user_account",
			input: VerifyRegistrationInput{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMocks: func(accountRepo *mocks.MockAccountRepository, verificationRepo *mocks.MockVerificationRegisterCodeRepository, cloudSvc *cloudinaryMocks.MockCloudinaryService) {
				// Valid verification record
				verificationRecord := &domain.RegistrationVerificationCode{
					ID:             "verification-id",
					Email:          "test@example.com",
					Name:           "Test User",
					HashedPassword: "hashed-password",
					Avatar:         "https://cloudinary.com/avatar.jpg",
					Code:           "123456",
					Type:           verification.VerificationCodeTypeRegister,
					ExpiresAt:      time.Now().Add(10 * time.Minute),
					CreatedAt:      time.Now(),
				}

				verificationRepo.On("GetVerificationCodeByEmail", ctx, "test@example.com").Return(verificationRecord, nil)

				// Create user account fails
				accountRepo.On("CreateUser", ctx, mock.MatchedBy(func(account *domain.Account) bool {
					return account.Email == "test@example.com" && account.Name == "Test User"
				})).Return(errors.New("database error"))
			},
			expectedResult: nil,
			expectedError:  "failed to create user account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAccountRepo := new(mocks.MockAccountRepository)
			mockVerificationRepo := new(mocks.MockVerificationRegisterCodeRepository)
			mockCloudService := new(cloudinaryMocks.MockCloudinaryService)

			// Setup mock expectations
			tt.setupMocks(mockAccountRepo, mockVerificationRepo, mockCloudService)

			// Create use case with mocks
			authUseCase := &authUseCase{
				accountRepository:              mockAccountRepo,
				verificationRegisterRepository: mockVerificationRepo,
				cloudstorage:                   mockCloudService,
			}

			// Execute
			result, err := authUseCase.VerifyRegistration(ctx, tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Name, result.Name)
				assert.Equal(t, tt.expectedResult.Email, result.Email)
				assert.Equal(t, tt.expectedResult.AvatarURL, result.AvatarURL)
				assert.NotEmpty(t, result.ID) // ID should be generated
			}

			// Verify all mock expectations were met
			mockAccountRepo.AssertExpectations(t)
			mockVerificationRepo.AssertExpectations(t)
			mockCloudService.AssertExpectations(t)
		})
	}
}
