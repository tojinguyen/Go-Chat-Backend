package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/upload"

	uploadHandler "gochat-backend/internal/handler/uploader"

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
