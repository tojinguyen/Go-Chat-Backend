package middleware

import (
	"gochat-backend/internal/handler"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *middleware) Authentication(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Authorization header is required")
		c.Abort()
		return
	}

	// Check if the header has the correct format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Invalid Authorization header format")
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := m.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		c.Abort()
		return
	}

	// Set user information in the context for use in handlers
	c.Set("userId", claims.UserId)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)

	c.Next()
}
