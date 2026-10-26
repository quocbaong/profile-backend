package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/service"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type ServiceHandler struct {
	serviceService service.ServiceService
}

func NewServiceHandler(serviceService service.ServiceService) *ServiceHandler {
	return &ServiceHandler{serviceService: serviceService}
}

// Public: Get all active services
func (h *ServiceHandler) GetPublicServices(c *gin.Context) {
	services, err := h.serviceService.GetAll(true)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve services", err)
		return
	}
	response.Success(c, http.StatusOK, "Services retrieved successfully", services)
}

// Admin: Get all services
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	services, err := h.serviceService.GetAll(false)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve services", err)
		return
	}
	response.Success(c, http.StatusOK, "Services retrieved successfully", services)
}

// Admin: Get by ID
func (h *ServiceHandler) GetServiceByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid service ID", err)
		return
	}

	srv, err := h.serviceService.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Service retrieved successfully", srv)
}

// Admin: Create service
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var srv models.ServiceItem
	if err := c.ShouldBindJSON(&srv); err != nil {
		response.BadRequest(c, "Invalid service input", err)
		return
	}

	if err := h.serviceService.Create(&srv); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusCreated, "Service created successfully", srv)
}

// Admin: Update service
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid service ID", err)
		return
	}

	var input models.ServiceItem
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid service input", err)
		return
	}

	updated, err := h.serviceService.Update(id, &input)
	if err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, "Service updated successfully", updated)
}

// Admin: Delete service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid service ID", err)
		return
	}

	if err := h.serviceService.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusOK, "Service deleted successfully", nil)
}
