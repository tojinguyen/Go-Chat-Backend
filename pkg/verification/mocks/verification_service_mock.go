package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockVerificationService struct {
	mock.Mock
}

func (m *MockVerificationService) GenerateCode() string {
	args := m.Called()
	return args.String(0)
}
