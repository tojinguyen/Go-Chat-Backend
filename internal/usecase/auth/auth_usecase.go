package auth

type AuthUseCase interface {
}

type authUseCase struct {
}

func NewAuthUseCase() AuthUseCase {
	return &authUseCase{}
}
