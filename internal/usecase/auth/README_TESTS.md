# Auth Use Case Tests

Đây là bộ test hoàn chỉnh cho auth use case của ứng dụng Realtime Chat.

## Các Test Cases Đã Implemented

### 1. Login Use Case (`login_usecase_test.go`)
- ✅ `successful_login` - Đăng nhập thành công
- ✅ `user_not_found` - Người dùng không tồn tại
- ✅ `user_account_is_nil` - Account trả về nil
- ✅ `invalid_password` - Mật khẩu không đúng
- ✅ `failed_to_generate_access_token` - Lỗi tạo access token
- ✅ `failed_to_generate_refresh_token` - Lỗi tạo refresh token

### 2. Register Use Case (`register_usecase_test.go`)
- ✅ `successful_registration_with_avatar` - Đăng ký thành công có avatar
- ✅ `successful_registration_without_avatar` - Đăng ký thành công không có avatar
- ✅ `email_already_exists` - Email đã tồn tại
- ✅ `error_checking_email_exists` - Lỗi kiểm tra email
- ✅ `delete_existing_verification_record` - Xóa verification record cũ
- ✅ `failed_to_upload_avatar` - Lỗi upload avatar
- ✅ `failed_to_send_email` - Lỗi gửi email

### 3. Verify Registration Use Case (`register_usecase_test.go`)
- ✅ `successful_verification` - Xác thực thành công
- ✅ `verification_record_not_found` - Không tìm thấy verification record
- ✅ `verification_code_expired` - Mã xác thực đã hết hạn
- ✅ `invalid_verification_code` - Mã xác thực không đúng
- ✅ `failed_to_create_user_account` - Lỗi tạo tài khoản người dùng

### 4. Verify Token Use Case (`verify_token_usecase_test.go`)
- ✅ `successful_token_verification` - Xác thực token thành công
- ✅ `token_without_bearer_prefix` - Token không có prefix "Bearer"
- ✅ `invalid_token` - Token không hợp lệ
- ✅ `user_not_found_in_database` - Không tìm thấy user trong DB
- ✅ `user_account_is_nil` - User account là nil

### 5. Refresh Token Use Case (`verify_token_usecase_test.go`)
- ✅ `successful_token_refresh` - Làm mới token thành công
- ✅ `invalid_refresh_token` - Refresh token không hợp lệ
- ✅ `user_not_found_during_refresh` - Không tìm thấy user khi refresh
- ✅ `user_account_is_nil_during_refresh` - User account nil khi refresh
- ✅ `failed_to_generate_access_token` - Lỗi tạo access token mới
- ✅ `failed_to_generate_refresh_token` - Lỗi tạo refresh token mới

## Mock Objects Đã Tạo

### Repository Mocks
- `MockAccountRepository` - Mock cho account repository
- `MockVerificationRegisterCodeRepository` - Mock cho verification repository

### Service Mocks
- `MockJwtService` - Mock cho JWT service
- `MockEmailService` - Mock cho email service
- `MockVerificationService` - Mock cho verification service
- `MockCloudinaryService` - Mock cho cloudinary service
- `MockRedisService` - Mock cho redis service

## Chạy Tests

### Chạy tất cả tests
```bash
go test ./internal/usecase/auth -v
```

### Chạy với coverage
```bash
go test ./internal/usecase/auth -v -cover
```

### Chạy với coverage detail
```bash
go test ./internal/usecase/auth -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Coverage

Tất cả các test cases đã PASS và cover:
- **Happy path scenarios** - Các trường hợp thành công
- **Error scenarios** - Các trường hợp lỗi
- **Edge cases** - Các trường hợp biên
- **Dependencies failures** - Lỗi từ các dependency

## Tiếp Theo

Có thể mở rộng với:
1. **Integration tests** - Test với database thật ✅ **ĐÃ HOÀN THÀNH**
2. **Performance tests** - Test hiệu năng
3. **Repository layer tests** - Test cho repository layer ✅ **ĐÃ HOÀN THÀNH** 
4. **Handler layer tests** - Test cho HTTP handlers
5. **WebSocket tests** - Test cho socket communication

## ✅ Integration Tests Đã Hoàn Thành

### Cấu trúc Integration Tests
```
tests/integration/
├── setup_test.go                    # Test setup và teardown
├── auth_integration_test.go         # Auth use case integration tests  
├── repository_integration_test.go   # Repository layer integration tests
└── redis_integration_test.go        # Redis operations integration tests
```

### Tính năng Integration Tests
- ✅ **Real Database Testing** - Test với MySQL thật
- ✅ **Real Redis Testing** - Test với Redis thật
- ✅ **GitHub Actions CI/CD** - Tự động chạy trên GitHub
- ✅ **Docker Support** - Setup database với Docker
- ✅ **Coverage Reports** - Báo cáo coverage riêng biệt
- ✅ **Automatic Cleanup** - Tự động dọn dẹp test data

### Chạy Integration Tests

#### Sử dụng scripts (Khuyến nghị):
```bash
# Windows PowerShell
.\scripts\run-integration-tests.ps1

# Linux/macOS
./scripts/run-integration-tests.sh
```

#### Sử dụng Makefile:
```bash
# Chạy integration tests với Docker
make test-integration-docker

# Chạy tất cả tests (unit + integration)
make test-all
```

#### Thủ công:
```bash
# 1. Start test databases
docker-compose -f docker-compose.test.yml up -d

# 2. Run migrations  
go run ./cmd/migrate/main.go

# 3. Run integration tests
go test -v -tags=integration ./tests/integration/...

# 4. Stop test databases
docker-compose -f docker-compose.test.yml down -v
```

### GitHub Actions Integration
- Tự động chạy trên push/PR đến `main`, `master`, `develop`
- Setup MySQL 8.0 và Redis services
- Chạy migrations và integration tests
- Upload coverage reports
- Workflow file: `.github/workflows/integration-tests.yml`

### Test Coverage Integration
- **Auth Use Case**: Login, Register, Verify với real database
- **Repository Layer**: Account, Verification CRUD operations 
- **Redis Operations**: Set/Get, Expiration, Cleanup
- **Database Constraints**: Foreign keys, unique constraints
- **Transaction Handling**: Commit/rollback scenarios

**Xem chi tiết**: [Integration Tests README](../../../tests/integration/README.md)
