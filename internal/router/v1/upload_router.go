package v1

import (
	uploadHandler "gochat-backend/internal/handler/upload"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/upload"

	"github.com/gin-gonic/gin"
)

func InitUploadRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	uploadUseCase upload.UploaderUseCase,
) {
	router.POST("/file-signature", middleware.Authentication, func(c *gin.Context) {
		uploadHandler.HandleUploadSignature(c, uploadUseCase)
	})
}
