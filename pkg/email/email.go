package email

import (
	"fmt"
	"gochat-backend/internal/config"
	"gochat-backend/pkg/verification"

	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendVerificationCode(toEmail string, code string, codeType verification.VerificationCodeType) error
}

type smtpEmailService struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTPEmailService(config *config.Environment) EmailService {
	return &smtpEmailService{
		Host:     config.EmailHost,
		Port:     config.EmailPort,
		Username: config.EmailUser,
		Password: config.EmailPass,
		From:     config.EmailName,
	}
}

func (s *smtpEmailService) SendVerificationCode(toEmail string, code string, codeType verification.VerificationCodeType) error {
	if !codeType.IsValid() {
		return fmt.Errorf("invalid verification code type: %s", codeType)
	}

	var subject string
	var bodyHTML string

	switch codeType {
	case verification.VerificationCodeTypeRegister:
		subject = "Welcome! Confirm your registration"
		bodyHTML = fmt.Sprintf(`<h2>Welcome!</h2><p>Your registration code is:</p><h1>%s</h1>`, code)
	case verification.VerificationCodeTypePasswordReset:
		subject = "Reset Your Password"
		bodyHTML = fmt.Sprintf(`<h2>Password Reset</h2><p>Use this code to reset your password:</p><h1>%s</h1>`, code)

	case verification.VerificationCodeTypeDeleteAccount:
		subject = "Confirm Account Deletion"
		bodyHTML = fmt.Sprintf(`<h2>Account Deletion</h2><p>Use this code to confirm deleting your account:</p><h1>%s</h1>`, code)

	default:
		subject = "Verification Code"
		bodyHTML = fmt.Sprintf(`<p>Your verification code is:</p><h1>%s</h1>`, code)

	}

	m := gomail.NewMessage()
	fromEmail := s.Username
	appPassword := s.Password

	m.SetHeader("From", fromEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)

	fullHTML := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial;">
				%s
				<p style="font-size: 12px; color: #888;">This code will expire in 10 minutes.</p>
			</body>
		</html>
	`, bodyHTML)

	m.SetBody("text/html", fullHTML)

	d := gomail.NewDialer(s.Host, s.Port, fromEmail, appPassword)

	return d.DialAndSend(m)
}
