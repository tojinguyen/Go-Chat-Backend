package auth

import (
	"context"
	"errors"
	"fmt"
	"gochat-backend/pkg/jwt"
	"log"
	"strings"
	"time"
)

// VerifyToken kiểm tra token truy cập và trả về thông tin người dùng nếu token hợp lệ
func (a *authUseCase) VerifyToken(ctx context.Context, token string) (*LoginOutput, error) {
	token = strings.TrimPrefix(token, "Bearer ")

	// Xác thực token truy cập
	claims, err := a.jwtService.ValidateAccessToken(token)
	if err != nil {
		log.Printf("Lỗi xác thực token: %v\n", err)
		return nil, fmt.Errorf("token không hợp lệ: %w", err)
	}

	// Lấy thông tin người dùng từ claims của token
	userId := claims.UserId
	email := claims.Email
	role := claims.Role

	// Lấy thông tin tài khoản từ repository
	account, err := a.accountRepository.FindById(ctx, userId)
	if err != nil {
		log.Printf("Lỗi tìm kiếm người dùng theo ID: %v\n", err)
		return nil, fmt.Errorf("không tìm thấy người dùng: %w", err)
	}

	if account == nil {
		return nil, errors.New("không tìm thấy người dùng")
	}

	// Trả về thông tin người dùng từ token
	return &LoginOutput{
		AccessToken:  token, // Trả về token đã được xác thực
		RefreshToken: "",    // Không bao gồm refresh token trong phản hồi xác thực
		UserId:       userId,
		Email:        email,
		FullName:     account.Name,
		Role:         role,
		AvatarUrl:    account.AvatarURL,
	}, nil
}

// RefreshToken kiểm tra refresh token và tạo cặp token mới
func (a *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*LoginOutput, error) {
	// Xác thực refresh token
	claims, err := a.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Printf("Lỗi xác thực refresh token: %v\n", err)
		return nil, fmt.Errorf("refresh token không hợp lệ: %w", err)
	}

	// Lấy thông tin người dùng từ claims
	userId := claims.UserId

	// Kiểm tra xem refresh token có tồn tại trong Redis không
	refreshTokenKey := fmt.Sprintf("refresh_token:%s", userId)
	var storedToken string
	err = a.redisService.Get(ctx, refreshTokenKey, &storedToken)
	if err != nil || storedToken != refreshToken {
		log.Printf("Refresh token không tìm thấy trong kho lưu trữ hoặc không khớp: %v\n", err)
		return nil, errors.New("refresh token đã bị thu hồi hoặc không hợp lệ")
	}

	// Lấy thông tin tài khoản để tạo token mới
	account, err := a.accountRepository.FindById(ctx, userId)
	if err != nil {
		log.Printf("Lỗi tìm kiếm người dùng theo ID: %v\n", err)
		return nil, fmt.Errorf("không tìm thấy người dùng: %w", err)
	}

	if account == nil {
		return nil, errors.New("không tìm thấy người dùng")
	}

	// Tạo token mới
	jwtInput := &jwt.GenerateTokenInput{
		UserId: account.Id,
		Email:  account.Email,
		Role:   claims.Role, // Giữ nguyên role từ token cũ
	}

	newAccessToken, err := a.jwtService.GenerateAccessToken(jwtInput)
	if err != nil {
		log.Printf("Lỗi tạo access token mới: %v\n", err)
		return nil, fmt.Errorf("không thể tạo access token: %w", err)
	}

	newRefreshToken, err := a.jwtService.GenerateRefreshToken(jwtInput)
	if err != nil {
		log.Printf("Lỗi tạo refresh token mới: %v\n", err)
		return nil, fmt.Errorf("không thể tạo refresh token: %w", err)
	}

	// Xóa refresh token cũ và lưu token mới
	err = a.redisService.Delete(ctx, refreshTokenKey)
	if err != nil {
		log.Printf("Lỗi xóa refresh token cũ: %v\n", err)
		// Tiếp tục xử lý, không cần dừng lại
	}

	// Lưu refresh token mới
	err = a.redisService.Set(ctx, refreshTokenKey, newRefreshToken, time.Duration(a.cfg.RefreshTokenExpireMinutes)*time.Minute)
	if err != nil {
		log.Printf("Lỗi lưu refresh token mới: %v\n", err)
		// Tiếp tục xử lý, không cần dừng lại
	}

	return &LoginOutput{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		UserId:       account.Id,
		Email:        account.Email,
		FullName:     account.Name,
		Role:         jwtInput.Role,
		AvatarUrl:    account.AvatarURL,
	}, nil
}
