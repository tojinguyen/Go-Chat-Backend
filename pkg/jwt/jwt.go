package jwt

import (
	"gochat-backend/config"
	"log"
	"time"

	errorConstants "gochat-backend/error"

	"github.com/golang-jwt/jwt"
)

type CustomJwtClaims struct {
	GenerateTokenInput
	jwt.StandardClaims
}

type GenerateTokenInput struct {
	UserId string
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
	now := time.Now().UTC()
	claims := &CustomJwtClaims{
		*input,
		jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: time.Now().UTC().Add(time.Duration(s.cfg.AccessTokenExpireMinutes) * time.Minute).Unix(),
			NotBefore: now.Unix(),
			Issuer:    "gochat-backend",
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
		// Kiểm tra thuật toán
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errorConstants.ErrTokenInvalid
		}
		return []byte(s.cfg.AccessTokenSecretKey), nil
	})

	if err != nil {
		log.Println("Error parsing token:", err)

		// Kiểm tra lỗi cụ thể
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				log.Println("Token expired")
				return nil, errorConstants.ErrTokenExpired
			}
		}

		return nil, errorConstants.ErrTokenInvalid
	}

	if !token.Valid {
		log.Println("Token is not valid")
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
