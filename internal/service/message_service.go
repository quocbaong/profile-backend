package service

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/repository"
)

type MessageService interface {
	GetAll() ([]models.ContactMessage, error)
	GetByID(id uint) (*models.ContactMessage, error)
	SubmitMessage(msg *models.ContactMessage) error
	MarkAsRead(id uint) error
	Delete(id uint) error
	GetUnreadCount() (int64, error)
}

type messageService struct {
	repo repository.MessageRepository
}

func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{repo: repo}
}

func (s *messageService) GetAll() ([]models.ContactMessage, error) {
	return s.repo.FindAll()
}

func (s *messageService) GetByID(id uint) (*models.ContactMessage, error) {
	msg, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, errors.New("message not found")
	}
	return msg, nil
}

func (s *messageService) SubmitMessage(msg *models.ContactMessage) error {
	if msg.Name == "" || msg.Email == "" || msg.Message == "" {
		return errors.New("name, email and message are required")
	}
	return s.repo.Create(msg)
}

func (s *messageService) MarkAsRead(id uint) error {
	return s.repo.MarkAsRead(id)
}

func (s *messageService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *messageService) GetUnreadCount() (int64, error) {
	return s.repo.CountUnread()
}
