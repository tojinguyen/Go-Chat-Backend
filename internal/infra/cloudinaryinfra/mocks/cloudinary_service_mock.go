package mocks

import (
	"gochat-backend/internal/infra/cloudinaryinfra"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

type MockCloudinaryService struct {
	mock.Mock
}

func (m *MockCloudinaryService) UploadAvatar(file *multipart.FileHeader, folderPath string) (string, error) {
	args := m.Called(file, folderPath)
	return args.String(0), args.Error(1)
}

func (m *MockCloudinaryService) MoveAvatar(avatarUrl string, fileName string) (string, error) {
	args := m.Called(avatarUrl, fileName)
	return args.String(0), args.Error(1)
}

func (m *MockCloudinaryService) GenerateUploadSignature(folderName string, resourceType string, optionalPublicID ...string) (*cloudinaryinfra.UploadSignatureResponse, error) {
	args := m.Called(folderName, resourceType, optionalPublicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cloudinaryinfra.UploadSignatureResponse), args.Error(1)
}
