package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/service"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type ProjectHandler struct {
	projectService service.ProjectService
}

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// Public endpoint: List active projects
func (h *ProjectHandler) GetPublicProjects(c *gin.Context) {
	projects, err := h.projectService.GetAll(true) // only active
	if err != nil {
		response.InternalServerError(c, "Failed to fetch projects", err)
		return
	}
	response.Success(c, http.StatusOK, "Projects retrieved", projects)
}

// Admin endpoint: List all projects (including inactive)
func (h *ProjectHandler) GetAllProjects(c *gin.Context) {
	projects, err := h.projectService.GetAll(false)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch projects", err)
		return
	}
	response.Success(c, http.StatusOK, "All projects retrieved", projects)
}

// Admin endpoint: Get project by ID
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID", err)
		return
	}

	project, err := h.projectService.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Project retrieved", project)
}

// Admin endpoint: Create project
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		response.BadRequest(c, "Invalid project input", err)
		return
	}

	if err := h.projectService.Create(&project); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusCreated, "Project created successfully", project)
}

// Admin endpoint: Update project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID", err)
		return
	}

	var input models.Project
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid project input", err)
		return
	}

	updated, err := h.projectService.Update(id, &input)
	if err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, "Project updated successfully", updated)
}

// Admin endpoint: Delete project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID", err)
		return
	}

	if err := h.projectService.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, "Project deleted successfully", nil)
}

func parseID(param string) (uint, error) {
	id64, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id64), nil
}
