package verification

type VerificationCodeType string

const (
	VerificationCodeTypeRegister VerificationCodeType = "REGISTER"

	VerificationCodeTypePasswordReset VerificationCodeType = "PASSWORD_RESET"

	VerificationCodeTypeDeleteAccount VerificationCodeType = "DELETE_ACCOUNT"
)

func (v VerificationCodeType) IsValid() bool {
	switch v {
	case
		VerificationCodeTypeRegister,
		VerificationCodeTypePasswordReset,
		VerificationCodeTypeDeleteAccount:
		return true
	}
	return false
}

func (v VerificationCodeType) String() string {
	return string(v)
}
