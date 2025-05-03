package handler

import (
	"gochat-backend/internal/handler"
	"gochat-backend/internal/usecase/profile"
	"log"
	"net/http"
	"strconv"

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
	name := c.Query("name")
	if name == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Name parameter is required")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	input := profile.SearchUsersInput{
		Name:  name,
		Page:  page,
		Limit: limit,
	}

	profiles, err := profileUseCase.SearchUsersByName(c.Request.Context(), input)
	if err != nil {
		log.Printf("Error searching users: %v\n", err)
		handler.SendErrorResponse(c, http.StatusInternalServerError, "Failed to search users")
		return
	}
	handler.SendSuccessResponse(c, http.StatusOK, "Users found", profiles)
}
