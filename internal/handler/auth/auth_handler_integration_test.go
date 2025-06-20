//go:build integration
// +build integration

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gochat-backend/config"
	domain "gochat-backend/internal/domain/auth"
	"gochat-backend/internal/handler"
	"gochat-backend/internal/infra/cloudinaryinfra/mocks"
	"gochat-backend/internal/infra/mysqlinfra"
	"gochat-backend/internal/infra/redisinfra"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/repository"
	"gochat-backend/internal/router"
	"gochat-backend/internal/usecase"
	jwtPkg "gochat-backend/pkg/jwt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	verificationPkg "gochat-backend/pkg/verification"
	verificationMocks "gochat-backend/pkg/verification/mocks"

	emailMocks "gochat-backend/pkg/email/mocks"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test server
func setupTestServer(t *testing.T, db *mysqlinfra.Database, redisService redisinfra.RedisService,
	mockCloudinary *mocks.MockCloudinaryService,
	mockEmail *emailMocks.MockEmailService,
	mockVerification *verificationMocks.MockVerificationService,
) *httptest.Server {
	cfg, err := config.Load() // Load từ .env.test
	require.NoError(t, err)

	// --- Dependencies ---
	// Real Repositories (vì đây là integration test cho handler, chúng ta muốn test cả repo)
	accountRepo := repository.NewAccountRepo(db, redisService)
	verificationRepo := repository.NewVerificationRepo(db)
	// ... other real repos if needed by register flow indirectly

	jwtService := jwtPkg.NewJwtService(cfg, redisService)

	deps := &usecase.SharedDependencies{
		Config:                   cfg,
		JwtService:               jwtService,
		EmailService:             mockEmail,        // Mocked
		VerificationService:      mockVerification, // Mocked
		AccountRepo:              accountRepo,
		VerificationRegisterRepo: verificationRepo,
		CloudinaryStorage:        mockCloudinary, // Mocked
		RedisService:             redisService,
		// ... other dependencies
	}
	useCaseContainer := usecase.NewUseCaseContainer(deps)
	mware := middleware.NewMiddleware(jwtService, nil, *cfg) // Logger có thể nil cho test đơn giản

	r := router.InitRouter(cfg, mware, useCaseContainer, deps)
	return httptest.NewServer(r)
}

// Helper để tạo multipart form request
func createRegisterRequest(t *testing.T, name, email, password string, avatarPath string) (*http.Request, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("name", name)
	_ = writer.WriteField("email", email)
	_ = writer.WriteField("password", password)

	if avatarPath != "" {
		file, err := os.Open(avatarPath)
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("avatar", filepath.Base(avatarPath))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
	}
	err := writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/auth/register", body)
	require.NoError(t, err)
	return req, writer
}

// Helper để tạo một file tạm cho avatar
func createTempAvatarFile(t *testing.T) (string, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "avatar-*.jpg")
	require.NoError(t, err)
	// Ghi một ít dữ liệu giả vào file
	_, err = tmpFile.Write([]byte("fake jpg data"))
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

func TestRegisterAPI(t *testing.T) {
	// Load .env.test
	err := godotenv.Load("../../../.env.test") // Điều chỉnh đường dẫn nếu cần
	require.NoError(t, err, "Failed to load .env.test file")

	cfg, err := config.Load()
	require.NoError(t, err)

	// Setup database connection for test
	db, err := mysqlinfra.ConnectMysql(cfg)
	require.NoError(t, err)
	defer db.Close()
	mysqlDB := mysqlinfra.NewMySqlDatabase(db)

	// Setup Redis connection for test
	redisService, err := redisinfra.NewRedisService(cfg)
	require.NoError(t, err)

	// Clear database tables before each test run in this suite
	clearTables := func() {
		_, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		require.NoError(t, err)
		tables := []string{"verification_codes", "users", "friendships", "friend_requests", "messages", "chat_room_members", "chat_rooms"} // Thêm các bảng khác nếu cần
		for _, table := range tables {
			_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table))
			if err != nil {
				// Một số table có thể không tồn tại nếu migration chưa chạy hết, hoặc không liên quan
				// Có thể log warning thay vì fail cứng ở đây tùy theo logic
				t.Logf("Warning: could not truncate table %s: %v", table, err)
			}
		}
		_, err = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		require.NoError(t, err)
		// Clear Redis (nếu cần, cẩn thận nếu Redis dùng chung)
		// Hoặc chỉ xóa các key liên quan đến test
		err = redisService.FlushAll(context.Background()) // Hoặc xóa theo prefix
		require.NoError(t, err)
	}

	// Create mock dependencies for external services
	mockCloudinary := new(mocks.MockCloudinaryService)
	mockEmail := new(emailMocks.MockEmailService)
	mockVerification := new(verificationMocks.MockVerificationService)

	// Setup and start the test server
	testServer := setupTestServer(t, mysqlDB, redisService, mockCloudinary, mockEmail, mockVerification)
	defer testServer.Close()

	// Create a temporary avatar file
	avatarPath, cleanupAvatar := createTempAvatarFile(t)
	defer cleanupAvatar()

	t.Run("Successful Registration", func(t *testing.T) {
		clearTables() // Dọn dẹp trước mỗi sub-test

		// --- Mock Expectations ---
		mockCloudinary.On("UploadAvatar", mock.AnythingOfType("*multipart.FileHeader"), "avatars/temp").Return("http://mockcloudinary.com/avatar.jpg", nil).Once()
		mockVerification.On("GenerateCode").Return("123456").Once()
		mockEmail.On("SendVerificationCode", "testuser@example.com", "123456", verificationPkg.VerificationCodeTypeRegister).Return(nil).Once()

		// --- Execute Request ---
		req, writer := createRegisterRequest(t, "Test User", "testuser@example.com", "Password123!", avatarPath)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		res, err := http.DefaultClient.Do(req) // Gửi request đến server test
		require.NoError(t, err)
		defer res.Body.Close()

		// --- Assertions ---
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		var responseBody handler.APIResponse
		err = json.NewDecoder(res.Body).Decode(&responseBody)
		require.NoError(t, err)

		assert.True(t, responseBody.Success)
		assert.Equal(t, "Registration successful! Please check your email for verification.", responseBody.Message)

		// Kiểm tra data trong response nếu có (ví dụ: thông tin user trả về)
		registerOutputData, ok := responseBody.Data.(map[string]interface{})
		require.True(t, ok, "Response data is not of expected type")
		assert.Equal(t, "Test User", registerOutputData["name"])
		assert.Equal(t, "testuser@example.com", registerOutputData["email"])
		assert.Equal(t, "http://mockcloudinary.com/avatar.jpg", registerOutputData["avatar_url"]) // URL từ mock
		assert.NotEmpty(t, registerOutputData["id"])                                              // ID của user được tạo bởi usecase

		// Kiểm tra database: verification_codes table
		var verificationRecord domain.RegistrationVerificationCode
		query := "SELECT id, email, name, hashed_password, avatar, code, type, expires_at FROM verification_codes WHERE email = ?"
		row := db.QueryRow(query, "testuser@example.com")
		err = row.Scan(
			&verificationRecord.ID, &verificationRecord.Email, &verificationRecord.Name,
			&verificationRecord.HashedPassword, &verificationRecord.Avatar, &verificationRecord.Code,
			&verificationRecord.Type, &verificationRecord.ExpiresAt,
		)
		require.NoError(t, err, "Verification record not found in DB or error scanning")
		assert.Equal(t, "Test User", verificationRecord.Name)
		assert.Equal(t, "http://mockcloudinary.com/avatar.jpg", verificationRecord.Avatar)
		assert.Equal(t, "123456", verificationRecord.Code)
		assert.True(t, verificationRecord.ExpiresAt.After(time.Now()))

		// Kiểm tra users table (chưa nên có user vì cần verify)
		var userCount int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "testuser@example.com").Scan(&userCount)
		require.NoError(t, err)
		assert.Equal(t, 0, userCount, "User should not be created before verification")

		// Verify mock calls
		mockCloudinary.AssertExpectations(t)
		mockEmail.AssertExpectations(t)
		mockVerification.AssertExpectations(t)
	})

	t.Run("Registration with Existing Email", func(t *testing.T) {
		clearTables()

		// Tạo user giả trong DB để test trường hợp email tồn tại
		_, err := db.Exec("INSERT INTO users (id, name, email, password, avatar_url) VALUES (?, ?, ?, ?, ?)",
			"existing-user-id", "Existing User", "existing@example.com", "hashedpassword", "url")
		require.NoError(t, err)

		// --- Mock Expectations (Cloudinary, Email, Verification không nên được gọi) ---

		// --- Execute Request ---
		req, writer := createRegisterRequest(t, "Another User", "existing@example.com", "Password123!", avatarPath)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		// --- Assertions ---
		assert.Equal(t, http.StatusBadRequest, res.StatusCode) // Hoặc code lỗi phù hợp

		var responseBody handler.APIResponse
		err = json.NewDecoder(res.Body).Decode(&responseBody)
		require.NoError(t, err)

		assert.False(t, responseBody.Success)
		assert.Contains(t, responseBody.Message, "email already exists")

		// Verify mocks (không có call nào được mong đợi)
		mockCloudinary.AssertNotCalled(t, "UploadAvatar", mock.Anything, mock.Anything)
		mockEmail.AssertNotCalled(t, "SendVerificationCode", mock.Anything, mock.Anything, mock.Anything)
		mockVerification.AssertNotCalled(t, "GenerateCode")
	})

	// Thêm các test case khác:
	// - Thiếu trường required (name, email, password)
	// - Email không hợp lệ
	// - Password quá yếu/ngắn
	// - Lỗi upload avatar (mock Cloudinary trả về lỗi)
	// - Lỗi gửi email (mock EmailService trả về lỗi)
}
