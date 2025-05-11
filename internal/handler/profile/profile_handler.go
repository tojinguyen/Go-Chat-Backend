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
// @Description Retrieves detailed profile information for a specific user by their ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} profile.ProfileOutput "Successfully retrieved user profile"
// @Failure 400 {object} handler.APIResponse "Invalid user ID format"
// @Failure 404 {object} handler.APIResponse "User not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /user/{id} [get]
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
// @Description Search for users by their name with pagination support
// @Tags User
// @Accept json
// @Produce json
// @Param name query string true "Name or partial name to search for" example("john")
// @Param page query int false "Page number for pagination results" default(1) minimum(1) example(1)
// @Param limit query int false "Number of results per page" default(10) minimum(1) maximum(100) example(20)
// @Success 200 {object} handler.APIResponse{data=profile.SearchUsersOutput} "List of users matching search criteria"
// @Failure 400 {object} handler.APIResponse "Missing required parameters or invalid pagination values"
// @Failure 401 {object} handler.APIResponse "Unauthorized access"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /user [get]
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
