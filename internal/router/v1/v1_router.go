package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/socket"
	"gochat-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func InitV1Router(
	r *gin.RouterGroup,
	middleware middleware.Middleware,
	useCaseContainer *usecase.UseCaseContainer,
	deps *usecase.SharedDependencies,
) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	{
		InitAuthRouter(r.Group("/auth"), middleware, useCaseContainer.Auth)
	}

	{
		InitUserRouter(r.Group("/users"), middleware, useCaseContainer.Profile)
	}

	{
		InitFriendRouter(r.Group("/friends"), middleware, useCaseContainer.Friend)
	}

	{
		InitChatRoomRouter(r.Group("/chat-rooms"), middleware, useCaseContainer.Chat)
	}

	{
		socketManager := socket.NewSocketManager(deps)
		InitWebSocketRouter(r.Group("/ws"), middleware, socketManager)
	}
}
