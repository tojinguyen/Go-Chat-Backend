package usecase

import (
	"gochat-backend/config"
	"gochat-backend/internal/infra/cloudinaryinfra"
	"gochat-backend/internal/infra/redisinfra"
	"gochat-backend/internal/repository"
	"gochat-backend/internal/usecase/auth"
	"gochat-backend/internal/usecase/chat"
	"gochat-backend/internal/usecase/friend"
	"gochat-backend/internal/usecase/profile"
	"gochat-backend/internal/usecase/status"
	uploader "gochat-backend/internal/usecase/upload"
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
	AccountRepo              repository.AccountRepository
	VerificationRegisterRepo repository.VerificationRegisterCodeRepository
	FriendShipRepo           repository.FriendShipRepository
	FriendRequestRepo        repository.FriendRequestRepository
	ChatRoomRepo             repository.ChatRoomRepository
	MessageRepo              repository.MessageRepository
	StatusRepo               repository.StatusRepository

	// Cloud Storage
	CloudinaryStorage cloudinaryinfra.CloudinaryService

	// Redis
	RedisService redisinfra.RedisService
}

type UseCaseContainer struct {
	Auth       auth.AuthUseCase
	Profile    profile.ProfileUseCase
	Friend     friend.FriendUseCase
	Chat       chat.ChatUseCase
	Uploader   uploader.UploaderUseCase
	UserStatus status.StatusUseCase
}

func NewUseCaseContainer(deps *SharedDependencies) *UseCaseContainer {
	return &UseCaseContainer{
		Auth: auth.NewAuthUseCase(
			deps.Config,
			deps.JwtService,
			deps.EmailService,
			deps.VerificationService,
			deps.AccountRepo,
			deps.VerificationRegisterRepo,
			deps.CloudinaryStorage,
			deps.RedisService,
		),
		Profile: profile.NewProfileUseCase(
			deps.AccountRepo,
		),
		Friend: friend.NewFriendUseCase(
			deps.FriendShipRepo,
			deps.FriendRequestRepo,
		),
		Chat: chat.NewChatUseCase(
			deps.ChatRoomRepo,
			deps.MessageRepo,
			deps.AccountRepo,
		),
		Uploader: uploader.NewUploaderUseCase(
			deps.CloudinaryStorage,
		),
		UserStatus: status.NewStatusUseCase(
			deps.StatusRepo,
			deps.Config,
		),
	}
}
