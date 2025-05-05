package verification

import (
	"fmt"
	"gochat-backend/config"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type VerificationService interface {
	GenerateCode() string
}

type verificationService struct {
	config *config.Environment
}

func NewVerificationService(config *config.Environment) VerificationService {
	return &verificationService{
		config: config,
	}
}

func (v *verificationService) GenerateCode() string {
	// Tạo một UUID ngẫu nhiên
	uuidObj := uuid.New()

	// Chuyển UUID thành chuỗi và loại bỏ dấu gạch ngang
	uuidStr := strings.ReplaceAll(uuidObj.String(), "-", "")

	// Nếu cần mã chỉ gồm số, chuyển các ký tự thành số
	numericCode := ""
	for i := 0; i < len(uuidStr) && len(numericCode) < int(v.config.VerificationCodeLength); i++ {
		// Chuyển từng ký tự hex thành giá trị số
		hexVal, _ := strconv.ParseInt(string(uuidStr[i]), 16, 64)
		numericCode += fmt.Sprintf("%d", hexVal%10)
	}

	// Đảm bảo độ dài của mã bằng với cấu hình
	if len(numericCode) > int(v.config.VerificationCodeLength) {
		numericCode = numericCode[:v.config.VerificationCodeLength]
	}

	return numericCode
}
