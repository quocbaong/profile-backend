package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/service"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type ProfileHandler struct {
	profileService service.ProfileService
}

func NewProfileHandler(profileService service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// Public & Admin: Lấy thông tin profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	profile, err := h.profileService.Get()
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Profile retrieved successfully", profile)
}

// Admin: Cập nhật thông tin profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var input models.ProfileSettings
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid input data", err)
		return
	}

	if err := h.profileService.Update(&input); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, "Profile updated successfully", input)
}
