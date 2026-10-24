package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/service"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type MessageHandler struct {
	messageService service.MessageService
}

func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// Public endpoint: Submit contact form message
func (h *MessageHandler) SubmitContact(c *gin.Context) {
	var msg models.ContactMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		response.BadRequest(c, "Invalid contact message format", err)
		return
	}

	if err := h.messageService.SubmitMessage(&msg); err != nil {
		response.BadRequest(c, err.Error(), err)
		return
	}

	response.Success(c, http.StatusCreated, "Message sent successfully", nil)
}

// Admin endpoint: List all contact messages
func (h *MessageHandler) GetAllMessages(c *gin.Context) {
	messages, err := h.messageService.GetAll()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch messages", err)
		return
	}
	response.Success(c, http.StatusOK, "Messages retrieved", messages)
}

// Admin endpoint: Mark message as read
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid message ID", err)
		return
	}

	if err := h.messageService.MarkAsRead(id); err != nil {
		response.InternalServerError(c, "Failed to update message", err)
		return
	}

	response.Success(c, http.StatusOK, "Message marked as read", nil)
}

// Admin endpoint: Delete message
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid message ID", err)
		return
	}

	if err := h.messageService.Delete(id); err != nil {
		response.InternalServerError(c, "Failed to delete message", err)
		return
	}

	response.Success(c, http.StatusOK, "Message deleted successfully", nil)
}

// Admin endpoint: Get unread count badge
func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	count, err := h.messageService.GetUnreadCount()
	if err != nil {
		response.InternalServerError(c, "Failed to get unread count", err)
		return
	}

	response.Success(c, http.StatusOK, "Unread count retrieved", gin.H{
		"unread_count": count,
	})
}
