package upload

import "gochat-backend/internal/infra/cloudinaryinfra"

type UploaderUseCase interface {
	GenerateUploadSignature(folderName string, resourceType string) (*cloudinaryinfra.UploadSignatureResponse, error)
}

type uploadUseCase struct {
	cloudinaryinfra cloudinaryinfra.CloudinaryService
}

func NewUploaderUseCase(cloudinaryinfra cloudinaryinfra.CloudinaryService) UploaderUseCase {
	return &uploadUseCase{
		cloudinaryinfra: cloudinaryinfra,
	}
}

func (u *uploadUseCase) GenerateUploadSignature(folderName string, resourceType string) (*cloudinaryinfra.UploadSignatureResponse, error) {
	return u.cloudinaryinfra.GenerateUploadSignature(folderName, resourceType)
}
