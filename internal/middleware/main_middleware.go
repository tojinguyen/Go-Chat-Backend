package middleware

import (
	"gochat-backend/internal/config"
	jwtPkg "gochat-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Middleware interface {
	RestLogger(context *gin.Context)
	Authentication(context *gin.Context)
}

type middleware struct {
	jwtService jwtPkg.JwtService
	logger     *logrus.Entry
	cfg        config.Environment
}

func NewMiddleware(
	jwtService jwtPkg.JwtService,
	logger *logrus.Entry,
	cfg config.Environment,
) Middleware {
	return &middleware{
		jwtService: jwtService,
		logger:     logger,
		cfg:        cfg,
	}
}
