package verification

import "gochat-backend/config"

type VerificationService interface {
	GenerateCode() string
}

type verificationService struct {
	config *config.Environment
}

func NewVerificationService(config *config.Environment) VerificationService {
	return &verificationService{
		config: config,
	}
}

func (v *verificationService) GenerateCode() string {
	code := make([]byte, v.config.VerificationCodeLength)
	for i := range code {
		code[i] = byte('0' + i%10) // Generate a digit (0-9)
	}
	return string(code)
}
