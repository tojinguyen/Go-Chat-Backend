package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRedisService struct {
	mock.Mock
}

func (m *MockRedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisService) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockRedisService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedisService) FlushAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
