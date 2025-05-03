package handler

import (
	"gochat-backend/internal/usecase/profile"

	"github.com/gin-gonic/gin"
)

func GetUserProfile(c *gin.Context, profileUseCase profile.ProfileUseCase) {
	userID := c.Param("id")
	profile, err := profileUseCase.GetUserProfile(c.Request.Context(), userID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, profile)
}

func SearchUsersByName(c *gin.Context, profileUseCase profile.ProfileUseCase) {
	query := c.Query("query")
	if query == "" {
		c.JSON(400, gin.H{"error": "Query parameter is required"})
		return
	}

	profiles, err := profileUseCase.SearchUsersByName(c.Request.Context(), profile.SearchUsersInput{Name: query})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, profiles)
}
