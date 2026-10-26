package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/service"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type SkillHandler struct {
	skillService service.SkillService
}

func NewSkillHandler(skillService service.SkillService) *SkillHandler {
	return &SkillHandler{skillService: skillService}
}

// Public: Get all active skills
func (h *SkillHandler) GetPublicSkills(c *gin.Context) {
	skills, err := h.skillService.GetAll(true)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve skills", err)
		return
	}
	response.Success(c, http.StatusOK, "Skills retrieved successfully", skills)
}

// Admin: Get all skills
func (h *SkillHandler) GetAllSkills(c *gin.Context) {
	skills, err := h.skillService.GetAll(false)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve skills", err)
		return
	}
	response.Success(c, http.StatusOK, "Skills retrieved successfully", skills)
}

// Admin: Get by ID
func (h *SkillHandler) GetSkillByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid skill ID", err)
		return
	}

	skill, err := h.skillService.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Skill retrieved successfully", skill)
}

// Admin: Create skill
func (h *SkillHandler) CreateSkill(c *gin.Context) {
	var skill models.SkillItem
	if err := c.ShouldBindJSON(&skill); err != nil {
		response.BadRequest(c, "Invalid skill input", err)
		return
	}

	if err := h.skillService.Create(&skill); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusCreated, "Skill created successfully", skill)
}

// Admin: Update skill
func (h *SkillHandler) UpdateSkill(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid skill ID", err)
		return
	}

	var input models.SkillItem
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid skill input", err)
		return
	}

	updated, err := h.skillService.Update(id, &input)
	if err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, "Skill updated successfully", updated)
}

// Admin: Delete skill
func (h *SkillHandler) DeleteSkill(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid skill ID", err)
		return
	}

	if err := h.skillService.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, "Skill deleted successfully", nil)
}
