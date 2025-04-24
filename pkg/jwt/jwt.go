package jwt

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


func GenerateAccessToken(email string) (string, error) {
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	if jwtKey == nil {
		return "", nil
	}

	timeExp := os.Getenv("JWT_EXPIRATION")

	if timeExp == "" {
		timeExp = "24" // Default expiration time in hours
	}

	expHours, err := strconv.Atoi(timeExp)
	if err != nil {
		return "", err
	}
	
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * time.Duration(expHours)).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}