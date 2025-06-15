package mocks

import (
	"context"
	"gochat-backend/pkg/jwt"

	"github.com/stretchr/testify/mock"
)

type MockJwtService struct {
	mock.Mock
}

func (m *MockJwtService) GenerateAccessToken(input *jwt.GenerateTokenInput) (string, error) {
	args := m.Called(input)
	return args.String(0), args.Error(1)
}

func (m *MockJwtService) GenerateRefreshToken(input *jwt.GenerateTokenInput) (string, error) {
	args := m.Called(input)
	return args.String(0), args.Error(1)
}

func (m *MockJwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*jwt.CustomJwtClaims, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomJwtClaims), args.Error(1)
}

func (m *MockJwtService) ValidateRefreshToken(tokenString string) (*jwt.CustomJwtClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomJwtClaims), args.Error(1)
}
