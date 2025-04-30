package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

func InitV1Router(
	r *gin.RouterGroup,
	middleware middleware.Middleware,
	authUseCase auth.AuthUseCase,
) {
	r.Use()
	{
		InitAuthRouter(r.Group("/auth"), middleware, authUseCase)
	}
}
