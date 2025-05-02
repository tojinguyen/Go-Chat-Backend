package config

var (
	StatusActive   = "on"
	StatusInactive = "off"
)

const (
	USER  = "USER"
	ADMIN = "ADMIN"
)

const (
	TypeSendEmail = "sendEmail"
)

type Constants struct {
	LoginFailLimit                            int `env:"LOGIN_FAIL_LIMIT,default=5"`
	LoginFailDurationMinutes                  int `env:"LOGIN_FAIL_DURATION_MINUTES,default=30"`
	ResetPasswordDurationHours                int `env:"RESET_PASSWORD_DURATION_HOURS,default=1"`
	ResendREquestResetPasswordDurationSeconds int `env:"RESEND_REQUEST_RESET_PASSWORD_DURATION_SECONDS,default=60"`
	SystemTimeOutSeconds                      int `env:"SYSTEM_TIME_OUT_SECONDS,default=60"`
}