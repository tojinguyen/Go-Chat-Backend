package mocks

import (
	"context"
	domain "gochat-backend/internal/domain/auth"

	"github.com/stretchr/testify/mock"
)

type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) CreateUser(ctx context.Context, account *domain.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepository) FindByEmail(ctx context.Context, email string) (*domain.Account, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) FindById(ctx context.Context, id string) (*domain.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccountRepository) UpdatePassword(ctx context.Context, id string, password string) error {
	args := m.Called(ctx, id, password)
	return args.Error(0)
}

func (m *MockAccountRepository) UpdateProfileInfo(ctx context.Context, account *domain.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepository) UpdateAvatar(ctx context.Context, id string, avatarURL string) error {
	args := m.Called(ctx, id, avatarURL)
	return args.Error(0)
}

func (m *MockAccountRepository) FindByName(ctx context.Context, name string, limit, offset int) ([]*domain.Account, error) {
	args := m.Called(ctx, name, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) CountByName(ctx context.Context, name string) (int, error) {
	args := m.Called(ctx, name)
	return args.Int(0), args.Error(1)
}
