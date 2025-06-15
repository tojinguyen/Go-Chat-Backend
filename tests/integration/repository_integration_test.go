//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountRepositoryIntegration(t *testing.T) {
	// Clean up before test
	cleanupTestData()

	accountRepo := repository.NewAccountRepo(MySQLService, RedisService)
	ctx := context.Background()

	t.Run("create_and_find_account_integration", func(t *testing.T) {
		// Create test account
		account := &domain.Account{
			Id:        "repo-test-123",
			Name:      "Repository Test User",
			Email:     "repotest@example.com",
			Password:  "hashedpassword123",
			AvatarURL: "https://example.com/avatar.jpg",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Test Create
		err := accountRepo.CreateUser(ctx, account)
		assert.NoError(t, err)

		// Test FindByEmail
		foundAccount, err := accountRepo.FindByEmail(ctx, "repotest@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, foundAccount)
		assert.Equal(t, "repo-test-123", foundAccount.Id)
		assert.Equal(t, "Repository Test User", foundAccount.Name)
		assert.Equal(t, "repotest@example.com", foundAccount.Email)
		assert.Equal(t, "hashedpassword123", foundAccount.Password)
		assert.Equal(t, "https://example.com/avatar.jpg", foundAccount.AvatarURL)

		// Test FindById
		foundAccountById, err := accountRepo.FindById(ctx, "repo-test-123")
		assert.NoError(t, err)
		assert.NotNil(t, foundAccountById)
		assert.Equal(t, foundAccount.Id, foundAccountById.Id)
		assert.Equal(t, foundAccount.Email, foundAccountById.Email)
	})

	t.Run("email_exists_check_integration", func(t *testing.T) {
		// Create test account
		account := &domain.Account{
			Id:        "exists-test-123",
			Name:      "Exists Test User",
			Email:     "existstest@example.com",
			Password:  "hashedpassword",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := accountRepo.CreateUser(ctx, account)
		require.NoError(t, err)

		// Test ExistsByEmail - should return true
		exists, err := accountRepo.ExistsByEmail(ctx, "existstest@example.com")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Test ExistsByEmail - should return false
		exists, err = accountRepo.ExistsByEmail(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("find_nonexistent_account_integration", func(t *testing.T) {
		// Test FindByEmail with non-existent email
		account, err := accountRepo.FindByEmail(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, account)

		// Test FindById with non-existent ID
		account, err = accountRepo.FindById(ctx, "nonexistent-id")
		assert.NoError(t, err)
		assert.Nil(t, account)
	})

	t.Run("duplicate_email_constraint_integration", func(t *testing.T) {
		// Create first account
		account1 := &domain.Account{
			Id:        "duplicate-test-1",
			Name:      "User 1",
			Email:     "duplicate@example.com",
			Password:  "password1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := accountRepo.CreateUser(ctx, account1)
		require.NoError(t, err)

		// Try to create second account with same email
		account2 := &domain.Account{
			Id:        "duplicate-test-2",
			Name:      "User 2",
			Email:     "duplicate@example.com", // Same email
			Password:  "password2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = accountRepo.CreateUser(ctx, account2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Duplicate entry")
	})
}

func TestVerificationRepositoryIntegration(t *testing.T) {
	// Clean up before test
	cleanupTestData()

	verificationRepo := repository.NewVerificationRepo(MySQLService)
	ctx := context.Background()
	t.Run("create_and_get_verification_integration", func(t *testing.T) {
		// Create test verification record
		verification := &domain.RegistrationVerificationCode{
			ID:             "verification-test-123",
			Name:           "Verification Test User",
			Email:          "verification@example.com",
			HashedPassword: "hashedpassword",
			Code:           "123456",
			ExpiresAt:      time.Now().Add(15 * time.Minute),
			CreatedAt:      time.Now(),
		}

		// Test Create
		err := verificationRepo.CreateVerificationCode(ctx, verification)
		assert.NoError(t, err)

		// Test GetByEmail
		foundVerification, err := verificationRepo.GetVerificationCodeByEmail(ctx, "verification@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, foundVerification)
		assert.Equal(t, "verification-test-123", foundVerification.ID)
		assert.Equal(t, "Verification Test User", foundVerification.Name)
		assert.Equal(t, "verification@example.com", foundVerification.Email)
		assert.Equal(t, "hashedpassword", foundVerification.HashedPassword)
		assert.Equal(t, "123456", foundVerification.Code)
		assert.Nil(t, foundVerification.VerifiedAt)
	})
	t.Run("update_verification_status_integration", func(t *testing.T) {
		// Create test verification record
		verification := &domain.RegistrationVerificationCode{
			ID:             "update-test-123",
			Name:           "Update Test User",
			Email:          "update@example.com",
			HashedPassword: "hashedpassword",
			Code:           "654321",
			ExpiresAt:      time.Now().Add(15 * time.Minute),
			CreatedAt:      time.Now(),
		}

		err := verificationRepo.CreateVerificationCode(ctx, verification)
		require.NoError(t, err)

		// Update verification status
		err = verificationRepo.UpdateVerificationStatus(ctx, "update-test-123", true)
		assert.NoError(t, err)

		// Verify the update
		foundVerification, err := verificationRepo.GetVerificationCodeByEmail(ctx, "update@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, foundVerification)
		assert.NotNil(t, foundVerification.VerifiedAt)
	})

	t.Run("delete_verification_integration", func(t *testing.T) {
		// Create test verification record
		verification := &domain.RegistrationVerificationCode{
			ID:             "delete-test-123",
			Name:           "Delete Test User",
			Email:          "delete@example.com",
			HashedPassword: "hashedpassword",
			Code:           "789012",
			ExpiresAt:      time.Now().Add(15 * time.Minute),
			CreatedAt:      time.Now(),
		}

		err := verificationRepo.CreateVerificationCode(ctx, verification)
		require.NoError(t, err)

		// Verify it exists
		foundVerification, err := verificationRepo.GetVerificationCodeByEmail(ctx, "delete@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, foundVerification)

		// Delete it
		err = verificationRepo.DeleteVerificationCode(ctx, "delete-test-123")
		assert.NoError(t, err)

		// Verify it's deleted
		foundVerification, err = verificationRepo.GetVerificationCodeByEmail(ctx, "delete@example.com")
		assert.NoError(t, err)
		assert.Nil(t, foundVerification)
	})

	t.Run("get_nonexistent_verification_integration", func(t *testing.T) {
		// Test GetByEmail with non-existent email
		verification, err := verificationRepo.GetVerificationCodeByEmail(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, verification)
	})
}
