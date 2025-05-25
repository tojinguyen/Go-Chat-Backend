package cloudinaryinfra

import (
	"context"
	"fmt"
	"gochat-backend/config"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/api"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type CloudinaryService interface {
	UploadAvatar(file *multipart.FileHeader, folderPath string) (string, error)
	MoveAvatar(avatarUrl string, fileName string) (string, error)
	GenerateUploadSignature(folderName string, optionalPublicID ...string) (*UploadSignatureResponse, error)
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

func (c *cloudinaryService) GenerateUploadSignature(folderName string, optionalPublicID ...string) (*UploadSignatureResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)

	paramsToSign := map[string]interface{}{
		"timestamp": timestamp,
	}

	if folderName != "" {
		paramsToSign["folder"] = folderName
	}

	var publicID string
	if len(optionalPublicID) > 0 && optionalPublicID[0] != "" {
		publicID = optionalPublicID[0]
	} else {
		// Generate a unique public_id if not provided
		publicID = uuid.New().String()
	}
	paramsToSign["public_id"] = publicID

	stringParams := make(map[string]string)
	stringParams["timestamp"] = timestamp
	stringParams["folder"] = folderName
	stringParams["public_id"] = publicID

	values := make(map[string][]string)
	for k, v := range stringParams {
		values[k] = []string{v}
	}

	signature, err := api.SignParameters(values, c.config.CloudinarySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign parameters: %w", err)
	}

	return &UploadSignatureResponse{
		Signature: signature,
		Timestamp: timestamp,
		APIKey:    c.config.CloudinaryKey,
		CloudName: c.config.CloudinaryName,
		Folder:    folderName,
		PublicID:  publicID,
	}, nil
}

func (c *cloudinaryService) UploadAvatar(file *multipart.FileHeader, folderPath string) (string, error) {
	ctx := context.Background()

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("cannot open file: %v", err)
	}
	defer src.Close()

	extension := filepath.Ext(file.Filename)
	uniqueFileName := uuid.New().String() + extension

	fmt.Printf("Uploading file %s to folder %s\n", file.Filename, folderPath)

	uploadParams := uploader.UploadParams{
		PublicID:       strings.TrimSuffix(uniqueFileName, extension),
		Folder:         folderPath,
		ResourceType:   "image",
		Transformation: "c_fill,g_face,w_500,h_500",
	}

	uploadResult, err := c.cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Cloudinary: %v", err)
	}

	if uploadResult.Error.Message != "" {
		return "", fmt.Errorf("failed to upload file to Cloudinary: %s", uploadResult.Error.Message)
	}

	fmt.Printf("Uploaded file to Cloudinary: %+v\n", uploadResult)

	if uploadResult == nil || uploadResult.SecureURL == "" {
		return "", fmt.Errorf("upload succeeded but returned empty URL")
	}

	return uploadResult.SecureURL, nil
}

func (c *cloudinaryService) MoveAvatar(avatarUrl string, fileName string) (string, error) {
	ctx := context.Background()

	publicIDWithPath := extractPublicIDFromURL(avatarUrl)

	fmt.Printf("Moving avatar from URL: %s\n", avatarUrl)
	fmt.Printf("Extracted public ID: %s\n", publicIDWithPath)

	extension := filepath.Ext(fileName)
	uniqueFileName := uuid.New().String() + extension

	uploadParams := uploader.UploadParams{
		PublicID:       strings.TrimSuffix(uniqueFileName, extension),
		Folder:         "avatars",
		ResourceType:   "image",
		Transformation: "w_500,h_500,c_fill,g_face",
	}

	uploadResult, err := c.cld.Upload.Upload(ctx, avatarUrl, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Cloudinary: %v", err)
	}

	// Kiểm tra lỗi và kết quả
	if uploadResult == nil || uploadResult.SecureURL == "" {
		return "", fmt.Errorf("upload succeeded but returned empty URL")
	}

	// Xóa file cũ sau khi đã upload xong
	if publicIDWithPath != "" {
		destroyParams := uploader.DestroyParams{
			PublicID:     publicIDWithPath,
			ResourceType: "image",
		}

		// Xóa file cũ
		destroyResult, err := c.cld.Upload.Destroy(ctx, destroyParams)
		if err != nil {
			// Chỉ log lỗi, không return vì file mới đã được tạo thành công
			fmt.Printf("Warning: Failed to delete old avatar: %v\n", err)
		} else {
			fmt.Printf("Deleted old avatar. Result: %v\n", destroyResult)
		}
	}

	return uploadResult.SecureURL, nil
}

// Hàm hỗ trợ để trích xuất public ID từ URL của Cloudinary
func extractPublicIDFromURL(url string) string {
	// URL Cloudinary thường có dạng: https://res.cloudinary.com/{cloud_name}/image/upload/v{version}/{folder}/{public_id}.{extension}

	// Tìm phần upload/ trong URL
	uploadIndex := strings.Index(url, "/upload/")
	if uploadIndex == -1 {
		return ""
	}

	// Lấy phần sau upload/
	pathAfterUpload := url[uploadIndex+8:] // +8 để bỏ qua "/upload/"

	// Loại bỏ phần version nếu có (v1234567890/)
	versionRegex := regexp.MustCompile(`^v\d+/`)
	pathAfterUpload = versionRegex.ReplaceAllString(pathAfterUpload, "")

	// Lấy phần trước extension
	extIndex := strings.LastIndex(pathAfterUpload, ".")
	if extIndex != -1 {
		pathAfterUpload = pathAfterUpload[:extIndex]
	}

	return pathAfterUpload
}
