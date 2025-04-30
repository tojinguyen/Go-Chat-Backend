package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func InitV1Router(
	r *gin.RouterGroup,
	middleware middleware.Middleware,
	useCaseContainer *usecase.UseCaseContainer,
) {
	r.Use()
	{
		InitAuthRouter(r.Group("/auth"), middleware, useCaseContainer.Auth)
	}
}
