//go:build integration
// +build integration

package handler_test

import "github.com/stretchr/testify/mock"

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(to, subject, htmlBody string) error {
	args := m.Called(to, subject, htmlBody)
	return args.Error(0)
}
