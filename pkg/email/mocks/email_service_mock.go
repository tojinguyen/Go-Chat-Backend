package mocks

import (
	"gochat-backend/pkg/verification"

	"github.com/stretchr/testify/mock"
)

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendVerificationCode(toEmail string, code string, codeType verification.VerificationCodeType) error {
	args := m.Called(toEmail, code, codeType)
	return args.Error(0)
}
