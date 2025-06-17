package main_test

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// TestMain sẽ được chạy trước tất cả các test trong package này.
func TestMain(m *testing.M) {
	log.Println("Setting up test environment...")

	// Xác định đường dẫn đến thư mục gốc của dự án
	// Điều này có thể cần điều chỉnh tùy thuộc vào vị trí file main_test.go
	// Giả sử file này đang ở internal/handler/auth, chúng ta cần đi lên 3 cấp
	err := godotenv.Load("../../.env.test") // Tải file .env.test
	if err != nil {
		log.Fatalf("Error loading .env.test file: %v", err)
	}

	log.Println("Test environment variables loaded from .env.test")

	// Khởi tạo và migrate database test

	// Chạy tất cả các test trong package
	exitVal := m.Run()
	// TODO: Thêm logic dọn dẹp sau khi test chạy xong (ví dụ: xóa database test)

	log.Println("Tearing down test environment...")
	os.Exit(exitVal)
}
