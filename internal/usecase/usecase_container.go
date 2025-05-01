package usecase

import (
	"gochat-backend/internal/config"
	cloudstorage "gochat-backend/internal/infra/cloudinaryinfra"
	"gochat-backend/internal/repository"
	"gochat-backend/internal/usecase/auth"
	"gochat-backend/pkg/email"
	"gochat-backend/pkg/jwt"
	"gochat-backend/pkg/verification"
)

type SharedDependencies struct {
	Config *config.Environment

	// Services
	JwtService          jwt.JwtService
	EmailService        email.EmailService
	VerificationService verification.VerificationService

	//Repositories
	AccountRepo repository.AccountRepository

	// Cloud Storage
	CloudStorage cloudstorage.CloudinaryService
}

type UseCaseContainer struct {
	Auth auth.AuthUseCase
}

func NewUseCaseContainer(deps *SharedDependencies) *UseCaseContainer {
	return &UseCaseContainer{
		Auth: auth.NewAuthUseCase(
			deps.Config,
			deps.JwtService,
			deps.EmailService,
			deps.VerificationService,
			deps.AccountRepo,
			deps.CloudStorage,
		),
	}
}
