package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/repository"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type DashboardHandler struct {
	projectRepo repository.ProjectRepository
	messageRepo repository.MessageRepository
}

func NewDashboardHandler(projectRepo repository.ProjectRepository, messageRepo repository.MessageRepository) *DashboardHandler {
	return &DashboardHandler{
		projectRepo: projectRepo,
		messageRepo: messageRepo,
	}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	allProjects, err := h.projectRepo.FindAll(false)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch projects count", err)
		return
	}

	activeCount := 0
	for _, p := range allProjects {
		if p.IsActive {
			activeCount++
		}
	}

	allMessages, err := h.messageRepo.FindAll()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch messages count", err)
		return
	}

	unreadCount, err := h.messageRepo.CountUnread()
	if err != nil {
		unreadCount = 0
	}

	response.Success(c, http.StatusOK, "Dashboard stats retrieved", gin.H{
		"total_projects":   len(allProjects),
		"active_projects":  activeCount,
		"total_messages":   len(allMessages),
		"unread_messages":  unreadCount,
	})
}
