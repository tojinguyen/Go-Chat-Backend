package cloudinaryinfra

import (
	"context"
	"fmt"
	"gochat-backend/internal/config"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type CloudinaryService interface {
	UploadAvatar(file *multipart.FileHeader, fileName string) (string, error)
	MoveAvatar(avatarUrl string, fileName string) (string, error)
}

type cloudinaryService struct {
	config *config.Environment
	cld    *cloudinary.Cloudinary
}

func NewCloudinaryService(cfg *config.Environment) (CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryName,
		cfg.CloudinaryKey,
		cfg.CloudinarySecret,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudinary client: %v", err)
	}

	return &cloudinaryService{
		config: cfg,
		cld:    cld,
	}, nil
}

func (c *cloudinaryService) UploadAvatar(file *multipart.FileHeader, fileName string) (string, error) {
	ctx := context.Background()

	extension := filepath.Ext(fileName)
	uniqueFileName := uuid.New().String() + extension

	uploadParams := uploader.UploadParams{
		PublicID:       strings.TrimSuffix(uniqueFileName, extension),
		Folder:         "avatars",
		ResourceType:   "image",
		Transformation: "w_500,h_500,c_fill,g_face",
	}

	uoloadResult, err := c.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Cloudinary: %v", err)
	}

	return uoloadResult.SecureURL, nil
}

func (c *cloudinaryService) MoveAvatar(avatarUrl string, fileName string) (string, error) {
	ctx := context.Background()

	extension := filepath.Ext(fileName)
	uniqueFileName := uuid.New().String() + extension

	uploadParams := uploader.UploadParams{
		PublicID:       strings.TrimSuffix(uniqueFileName, extension),
		Folder:         "avatars",
		ResourceType:   "image",
		Transformation: "w_500,h_500,c_fill,g_face",
	}

	uoloadResult, err := c.cld.Upload.Upload(ctx, avatarUrl, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Cloudinary: %v", err)
	}

	return uoloadResult.SecureURL, nil
}
