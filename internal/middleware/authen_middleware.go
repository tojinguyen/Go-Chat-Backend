package middleware

import (
	"gochat-backend/internal/handler"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *middleware) Authentication(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	var tokenString string

	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Nếu không có header, thử lấy token từ query string (dùng cho WebSocket)
		tokenString = c.Query("token")
	}

	if tokenString == "" {
		log.Println("Authorization token is missing")
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Authorization token is required")
		c.Abort()
		return
	}

	claims, err := m.jwtService.ValidateAccessToken(c, tokenString)
	if err != nil {
		log.Println("Invalid token:", err)
		handler.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		c.Abort()
		return
	}

	// Set user information in the context for use in handlers
	c.Set("userId", claims.UserId)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)

	log.Println("User ID from token:", claims.UserId)

	c.Next()
}
