package jwt

import (
	"gochat-backend/internal/config"
	"time"

	errorConstants "gochat-backend/internal/error"

	"github.com/golang-jwt/jwt"
)

type CustomJwtClaims struct {
	GenerateTokenInput
	jwt.StandardClaims
}

type GenerateTokenInput struct {
	UserId int
	Email  string
	Role   string
}

type JwtService interface {
	GenerateAccessToken(input *GenerateTokenInput) (string, error)
	GenerateRefreshToken(input *GenerateTokenInput) (string, error)
	ValidateAccessToken(tokenString string) (*CustomJwtClaims, error)
	ValidateRefreshToken(tokenString string) (*CustomJwtClaims, error)
}

type jwtService struct {
	cfg *config.Environment
}

func NewJwtService(cfg *config.Environment) JwtService {
	return &jwtService{
		cfg: cfg,
	}
}

func (s *jwtService) GenerateAccessToken(input *GenerateTokenInput) (string, error) {
	claims := &CustomJwtClaims{
		*input,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(s.cfg.AccessTokenExpireMinutes) * time.Minute).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.AccessTokenSecretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *jwtService) GenerateRefreshToken(input *GenerateTokenInput) (string, error) {
	claims := &CustomJwtClaims{
		*input,
		jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Add(time.Duration(s.cfg.RefreshTokenExpireMinutes) * time.Minute).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.RefreshTokenSecretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *jwtService) ValidateAccessToken(tokenString string) (*CustomJwtClaims, error) {
	claims := &CustomJwtClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.AccessTokenSecretKey), nil
	})

	if err != nil {
		v, _ := err.(*jwt.ValidationError)

		if v.Errors == jwt.ValidationErrorExpired {
			return nil, errorConstants.ErrTokenExpired
		}

		return nil, errorConstants.ErrTokenInvalid
	}

	if !token.Valid {
		return nil, errorConstants.ErrTokenInvalid
	}

	return claims, nil
}

func (s *jwtService) ValidateRefreshToken(tokenString string) (*CustomJwtClaims, error) {
	claims := &CustomJwtClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.RefreshTokenSecretKey), nil
	})

	if err != nil {
		v, _ := err.(*jwt.ValidationError)

		if v.Errors == jwt.ValidationErrorExpired {
			return nil, errorConstants.ErrTokenExpired
		}

		return nil, errorConstants.ErrTokenInvalid
	}

	if !token.Valid {
		return nil, errorConstants.ErrTokenInvalid
	}

	return claims, nil
}
