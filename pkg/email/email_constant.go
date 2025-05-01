package email

type VerificationCodeType string

const (
	VerificationCodeTypeRegister VerificationCodeType = "REGISTER"

	VerificationCodeTypePasswordReset VerificationCodeType = "PASSWORD_RESET"

	VerificationCodeTypeDeleteAccount VerificationCodeType = "DELETE_ACCOUNT"
)
