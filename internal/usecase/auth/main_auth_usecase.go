package auth

import (
	"context"
	"gochat-backend/config"
	cloudstorage "gochat-backend/internal/infra/cloudinaryinfra"
	"gochat-backend/internal/infra/redisinfra"
	"gochat-backend/internal/repository"
	"gochat-backend/pkg/email"
	"gochat-backend/pkg/jwt"
	"gochat-backend/pkg/verification"
)

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	VerifyRegistration(ctx context.Context, input VerifyRegistrationInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
}

type authUseCase struct {
	cfg                            *config.Environment
	jwtService                     jwt.JwtService
	emailService                   email.EmailService
	verificationService            verification.VerificationService
	accountRepository              repository.AccountRepository
	verificationRegisterRepository repository.VerificationRegisterCodeRepository

	cloudstorage cloudstorage.CloudinaryService
	redisService redisinfra.RedisService
}

func NewAuthUseCase(
	cfg *config.Environment,
	jwtService jwt.JwtService,
	emailService email.EmailService,
	verificationService verification.VerificationService,
	accountRepository repository.AccountRepository,
	verificationRegisterRepository repository.VerificationRegisterCodeRepository,
	cloudstorage cloudstorage.CloudinaryService,
	redisService redisinfra.RedisService,
) AuthUseCase {
	return &authUseCase{
		cfg:                            cfg,
		jwtService:                     jwtService,
		emailService:                   emailService,
		verificationService:            verificationService,
		verificationRegisterRepository: verificationRegisterRepository,
		accountRepository:              accountRepository,
		cloudstorage:                   cloudstorage,
		redisService:                   redisService,
	}
}
