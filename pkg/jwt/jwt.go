package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gochat-backend/config"
	"gochat-backend/internal/infra/redisinfra"
	"log"
	"time"

	errorConstants "gochat-backend/error"

	"github.com/golang-jwt/jwt"
)

const (
	tokenClaimsCacheKeyPrefix = "token_claims_v2:" // Thêm _v2 để tránh trùng key cũ nếu có
	// TTL nên ngắn hơn hoặc bằng AccessTokenExpireMinutes một chút để đảm bảo
	// cache hết hạn cùng lúc hoặc trước token.
	// Ví dụ: nếu token hết hạn sau 60 phút, cache có thể hết hạn sau 59 phút.
	tokenClaimsCacheTTLExpiryFactor = 0.98
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
	ValidateAccessToken(ctx context.Context, tokenString string) (*CustomJwtClaims, error)
	ValidateRefreshToken(tokenString string) (*CustomJwtClaims, error)
}

type jwtService struct {
	cfg          *config.Environment
	redisService redisinfra.RedisService
}

func NewJwtService(cfg *config.Environment, redisService redisinfra.RedisService) JwtService {
	return &jwtService{
		cfg:          cfg,
		redisService: redisService,
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

func (s *jwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*CustomJwtClaims, error) {
	cacheKey := s.generateTokenCacheKey(tokenString)
	var cachedClaims CustomJwtClaims

	// 1. Try to get claims from cache
	if s.redisService != nil {
		if err := s.redisService.Get(ctx, cacheKey, &cachedClaims); err == nil {
			// Kiểm tra xem claims đã cache có còn hợp lệ không (phòng trường hợp TTL của cache > TTL token)
			// Dù StandardClaims.Valid() đã kiểm tra ExpiresAt, nhưng cẩn thận hơn.
			if time.Now().UTC().Unix() < cachedClaims.StandardClaims.ExpiresAt {
				log.Println("Access token claims cache hit:", tokenString)
				return &cachedClaims, nil
			}
			// Nếu claims đã cache hết hạn, xóa nó đi
			_ = s.redisService.Delete(ctx, cacheKey)
		}
	}

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

	// 3. Lưu vào Cache
	if s.redisService != nil {
		// Tính TTL cho cache dựa trên thời gian hết hạn của token
		ttl := time.Until(time.Unix(claims.ExpiresAt, 0))
		if ttl > 0 { // Chỉ cache nếu token còn hạn
			// Giảm TTL đi một chút để đảm bảo cache hết hạn trước hoặc cùng lúc với token
			adjustedTTL := time.Duration(float64(ttl) * tokenClaimsCacheTTLExpiryFactor)
			if adjustedTTL > 1*time.Second { // Đảm bảo TTL > 0
				if err := s.redisService.Set(ctx, cacheKey, claims, adjustedTTL); err != nil {
					log.Printf("Warning: Failed to cache access token claims (token: %s): %v\n", tokenString, err)
				} else {
					log.Println("Access token claims cached:", tokenString)
				}
			}
		}
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

func (s *jwtService) generateTokenCacheKey(tokenString string) string {
	hasher := sha256.New()
	hasher.Write([]byte(tokenString))
	return fmt.Sprintf("%s%s", tokenClaimsCacheKeyPrefix, hex.EncodeToString(hasher.Sum(nil)))
}
