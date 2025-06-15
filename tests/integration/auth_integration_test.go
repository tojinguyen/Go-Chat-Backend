//go:build integration

package integration

import (
	"context"
	"mime/multipart"
	"testing"
	"time"

	"gochat-backend/config"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/infra/cloudinaryinfra"
	"gochat-backend/internal/repository"
	"gochat-backend/internal/usecase/auth"
	"gochat-backend/pkg/jwt"
	"gochat-backend/pkg/verification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCaseIntegration_Login(t *testing.T) {
	// Clean up before test
	cleanupTestData() // Setup repositories and services
	accountRepo := repository.NewAccountRepo(MySQLService, RedisService)
	jwtService := jwt.NewJwtService(TestEnv, RedisService)
	redisService := RedisService

	// Create auth use case
	authUseCase := auth.NewAuthUseCase(
		TestEnv,
		jwtService,
		nil, // email service not needed for login
		nil, // verification service not needed for login
		accountRepo,
		nil, // verification repo not needed for login
		nil, // cloudinary service not needed for login
		redisService,
	)

	ctx := context.Background()

	// Test data setup
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Insert test user directly into database
	testUser := &domain.Account{
		Id:        "test-user-123",
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		AvatarURL: "https://example.com/avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = accountRepo.CreateUser(ctx, testUser)
	require.NoError(t, err)

	t.Run("successful_login_integration", func(t *testing.T) {
		input := auth.LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}

		result, err := authUseCase.Login(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "test-user-123", result.UserId)
		assert.Equal(t, "test@example.com", result.Email)
		assert.Equal(t, "Test User", result.FullName)
		assert.Equal(t, config.USER, result.Role)
		assert.Equal(t, "https://example.com/avatar.jpg", result.AvatarUrl)
		// Verify refresh token is stored in Redis
		var storedToken string
		err = RedisService.Get(ctx, "refresh_token:test-user-123", &storedToken)
		assert.NoError(t, err)
		assert.Equal(t, result.RefreshToken, storedToken)
	})

	t.Run("user_not_found_integration", func(t *testing.T) {
		input := auth.LoginInput{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		result, err := authUseCase.Login(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("invalid_password_integration", func(t *testing.T) {
		input := auth.LoginInput{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		result, err := authUseCase.Login(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid password")
	})
}

func TestAuthUseCaseIntegration_Register(t *testing.T) {
	// Clean up before test
	cleanupTestData() // Setup repositories and services
	accountRepo := repository.NewAccountRepo(MySQLService, RedisService)
	verificationRepo := repository.NewVerificationRepo(MySQLService)
	jwtService := jwt.NewJwtService(TestEnv, RedisService)
	emailService := &MockEmailService{} // Mock email service for integration test
	verificationService := verification.NewVerificationService(TestEnv)
	cloudinaryService := &MockCloudinaryService{} // Mock cloudinary service for integration test

	// Create auth use case
	authUseCase := auth.NewAuthUseCase(
		TestEnv,
		jwtService,
		emailService,
		verificationService,
		accountRepo,
		verificationRepo,
		cloudinaryService,
		RedisService,
	)

	ctx := context.Background()

	t.Run("successful_registration_integration", func(t *testing.T) {
		input := auth.RegisterInput{
			Name:     "New User",
			Email:    "newuser@example.com",
			Password: "password123",
		}

		result, err := authUseCase.Register(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New User", result.Name)
		assert.Equal(t, "newuser@example.com", result.Email)
		assert.NotEmpty(t, result.ID)
		// Verify verification code was created in database
		verificationCode, err := verificationRepo.GetVerificationCodeByEmail(ctx, "newuser@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, verificationCode)
		assert.Equal(t, "newuser@example.com", verificationCode.Email)
		assert.NotEmpty(t, verificationCode.Code)
		assert.Nil(t, verificationCode.VerifiedAt)
	})
	t.Run("email_already_exists_integration", func(t *testing.T) {
		// First, create a user
		testUser := &domain.Account{
			Id:        "existing-user-123",
			Name:      "Existing User",
			Email:     "existing@example.com",
			Password:  "hashedpassword",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := accountRepo.CreateUser(ctx, testUser)
		require.NoError(t, err)

		// Try to register with the same email
		input := auth.RegisterInput{
			Name:     "Another User",
			Email:    "existing@example.com",
			Password: "password123",
		}

		result, err := authUseCase.Register(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "email already exists")
	})
}

func TestAuthUseCaseIntegration_VerifyRegistration(t *testing.T) {
	// Clean up before test
	cleanupTestData() // Setup repositories and services
	accountRepo := repository.NewAccountRepo(MySQLService, RedisService)
	verificationRepo := repository.NewVerificationRepo(MySQLService)
	jwtService := jwt.NewJwtService(TestEnv, RedisService)

	// Create mock services
	emailService := &MockEmailService{}
	verificationService := verification.NewVerificationService(TestEnv)
	cloudinaryService := &MockCloudinaryService{}

	// Create auth use case
	authUseCase := auth.NewAuthUseCase(
		TestEnv,
		jwtService,
		emailService,
		verificationService,
		accountRepo,
		verificationRepo,
		cloudinaryService,
		RedisService,
	)

	ctx := context.Background()
	t.Run("successful_verification_integration", func(t *testing.T) {
		// Setup test verification record
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		verificationRecord := &domain.RegistrationVerificationCode{
			ID:             "verification-123",
			Name:           "Test User",
			Email:          "verify@example.com",
			HashedPassword: string(hashedPassword),
			Code:           "123456",
			ExpiresAt:      time.Now().Add(15 * time.Minute),
			CreatedAt:      time.Now(),
		}

		err = verificationRepo.CreateVerificationCode(ctx, verificationRecord)
		require.NoError(t, err)

		// Test verification
		input := auth.VerifyRegistrationInput{
			Email: "verify@example.com",
			Code:  "123456",
		}

		result, err := authUseCase.VerifyRegistration(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "verify@example.com", result.Email)
		assert.Equal(t, "Test User", result.Name)
		assert.NotEmpty(t, result.ID)

		// Verify user account was created
		account, err := accountRepo.FindByEmail(ctx, "verify@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, "Test User", account.Name)
		assert.Equal(t, "verify@example.com", account.Email)

		// Verify verification record was marked as verified
		verificationCode, err := verificationRepo.GetVerificationCodeByEmail(ctx, "verify@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, verificationCode.VerifiedAt)
	})
	t.Run("verification_code_not_found_integration", func(t *testing.T) {
		input := auth.VerifyRegistrationInput{
			Email: "notfound@example.com",
			Code:  "123456",
		}

		result, err := authUseCase.VerifyRegistration(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "verification record not found")
	})

	t.Run("invalid_verification_code_integration", func(t *testing.T) {
		// Setup test verification record
		verificationRecord := &domain.RegistrationVerificationCode{
			ID:             "verification-invalid-123",
			Name:           "Test User",
			Email:          "invalid@example.com",
			HashedPassword: "hashedpassword",
			Code:           "123456",
			ExpiresAt:      time.Now().Add(15 * time.Minute),
			CreatedAt:      time.Now(),
		}

		err := verificationRepo.CreateVerificationCode(ctx, verificationRecord)
		require.NoError(t, err)

		// Test with wrong code
		input := auth.VerifyRegistrationInput{
			Email: "invalid@example.com",
			Code:  "wrong_code",
		}

		result, err := authUseCase.VerifyRegistration(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid verification code")
	})

	t.Run("expired_verification_code_integration", func(t *testing.T) {
		// Setup expired verification record
		verificationRecord := &domain.RegistrationVerificationCode{
			ID:             "verification-expired-123",
			Name:           "Test User",
			Email:          "expired@example.com",
			HashedPassword: "hashedpassword",
			Code:           "123456",
			ExpiresAt:      time.Now().Add(-1 * time.Minute), // Expired
			CreatedAt:      time.Now(),
		}

		err := verificationRepo.CreateVerificationCode(ctx, verificationRecord)
		require.NoError(t, err)

		// Test verification
		input := auth.VerifyRegistrationInput{
			Email: "expired@example.com",
			Code:  "123456",
		}

		result, err := authUseCase.VerifyRegistration(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "verification code has expired")
	})
}

// Mock services for integration tests
type MockEmailService struct{}

func (m *MockEmailService) SendVerificationCode(toEmail string, code string, codeType verification.VerificationCodeType) error {
	return nil
}

type MockCloudinaryService struct{}

func (m *MockCloudinaryService) UploadAvatar(file *multipart.FileHeader, folderPath string) (string, error) {
	return "https://example.com/avatar.jpg", nil
}

func (m *MockCloudinaryService) MoveAvatar(avatarUrl string, fileName string) (string, error) {
	return "https://example.com/avatar.jpg", nil
}

func (m *MockCloudinaryService) GenerateUploadSignature(folderName string, resourceType string, optionalPublicID ...string) (*cloudinaryinfra.UploadSignatureResponse, error) {
	return &cloudinaryinfra.UploadSignatureResponse{}, nil
}
