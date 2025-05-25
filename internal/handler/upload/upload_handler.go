package handler

import (
	"gochat-backend/internal/handler"
	"gochat-backend/internal/usecase/upload"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleUploadSignature godoc
// @Summary Get Cloudinary Upload Signature
// @Description Generates a signature and parameters for direct client-side upload to Cloudinary for chat files.
// @Tags Upload
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} handler.APIResponse{data=cloudinaryinfra.UploadSignatureResponse} "Successfully generated signature"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /api/v1/chat/upload-signature [post]
func HandleUploadSignature(c *gin.Context, upload upload.UploaderUseCase) {
	_, exists := c.Get("userId")
	if !exists {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	folderName := "chat_files"

	signatureResponse, err := upload.GenerateUploadSignature(folderName)
	if err != nil {
		log.Printf("Error generating upload signature: %v", err)
		handler.SendErrorResponse(c, http.StatusInternalServerError, "Failed to generate upload signature")
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Signature generated successfully", signatureResponse)
}
