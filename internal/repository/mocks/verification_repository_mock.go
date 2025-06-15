package mocks

import (
	"context"
	domain "gochat-backend/internal/domain/auth"

	"github.com/stretchr/testify/mock"
)

type MockVerificationRegisterCodeRepository struct {
	mock.Mock
}

func (m *MockVerificationRegisterCodeRepository) CreateVerificationCode(ctx context.Context, code *domain.RegistrationVerificationCode) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockVerificationRegisterCodeRepository) GetVerificationCodeByID(ctx context.Context, id string) (*domain.RegistrationVerificationCode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RegistrationVerificationCode), args.Error(1)
}

func (m *MockVerificationRegisterCodeRepository) GetVerificationCodeByEmail(ctx context.Context, email string) (*domain.RegistrationVerificationCode, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RegistrationVerificationCode), args.Error(1)
}

func (m *MockVerificationRegisterCodeRepository) VerifyCode(ctx context.Context, id string, code string) error {
	args := m.Called(ctx, id, code)
	return args.Error(0)
}

func (m *MockVerificationRegisterCodeRepository) UpdateVerificationStatus(ctx context.Context, id string, verified bool) error {
	args := m.Called(ctx, id, verified)
	return args.Error(0)
}

func (m *MockVerificationRegisterCodeRepository) DeleteVerificationCode(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
