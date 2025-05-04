package handler

import (
	"gochat-backend/internal/handler"
	"gochat-backend/internal/usecase/profile"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserProfile godoc
// @Summary Get user profile details
// @Description Fetch a user's profile information by their ID
// @Tags Profile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} profile.ProfileOutput
// @Failure 404 {object} handler.APIResponse "User not found"
// @Failure 500 {object} handler.APIResponse "Server error"
// @Security BearerAuth
// @Router /profile/users/{id} [get]
func GetUserProfile(c *gin.Context, profileUseCase profile.ProfileUseCase) {
	userID := c.Param("id")
	profile, err := profileUseCase.GetUserProfile(c.Request.Context(), userID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, profile)
}

// SearchUsersByName godoc
// @Summary Search users by name
// @Description Search for users with pagination by their name
// @Tags Profile
// @Accept json
// @Produce json
// @Param name query string true "Name to search for"
// @Param page query int false "Page number (default: 1)" default(1) minimum(1)
// @Param limit query int false "Number of items per page (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Success 200 {object} handler.APIResponse{data=profile.SearchUsersOutput}
// @Failure 400 {object} handler.APIResponse "Name parameter is required"
// @Failure 500 {object} handler.APIResponse "Failed to search users"
// @Security BearerAuth
// @Router /profile/users [get]
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
