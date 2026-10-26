package service

import (
	"errors"
	"log"
	"net/mail"
	"regexp"
	"strings"

	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/repository"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type MessageService interface {
	GetAll() ([]models.ContactMessage, error)
	GetByID(id uint) (*models.ContactMessage, error)
	SubmitMessage(msg *models.ContactMessage) error
	MarkAsRead(id uint) error
	Delete(id uint) error
	GetUnreadCount() (int64, error)
}

type messageService struct {
	repo         repository.MessageRepository
	emailService EmailService
}

func NewMessageService(repo repository.MessageRepository, emailService EmailService) MessageService {
	return &messageService{
		repo:         repo,
		emailService: emailService,
	}
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
	if err := s.validateAndSanitize(msg); err != nil {
		return err
	}

	// 1. Persist to database first so no message is ever lost
	if err := s.repo.Create(msg); err != nil {
		return err
	}

	// 2. Non-blocking asynchronous email delivery via background goroutine
	if s.emailService != nil {
		msgCopy := *msg // Safe copy to avoid data race
		go func(m models.ContactMessage) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[EmailService PANIC RECOVER] %v", r)
				}
			}()

			// Send alert notification to portfolio admin
			if err := s.emailService.SendAdminNotification(&m); err != nil {
				log.Printf("[EmailService ERROR] Failed to send admin notification: %v", err)
			}

			// Send automated polite acknowledgement to visitor
			if err := s.emailService.SendAutoReply(&m); err != nil {
				log.Printf("[EmailService ERROR] Failed to send auto-reply to %s: %v", m.Email, err)
			}
		}(msgCopy)
	}

	return nil
}

func (s *messageService) validateAndSanitize(msg *models.ContactMessage) error {
	if msg == nil {
		return errors.New("message payload cannot be empty")
	}

	// Sanitize against Header Injection (\r, \n) and trim excess spaces
	msg.Name = strings.TrimSpace(sanitizeHeader(msg.Name))
	msg.LastName = strings.TrimSpace(sanitizeHeader(msg.LastName))
	msg.Email = strings.TrimSpace(sanitizeHeader(msg.Email))
	msg.Subject = strings.TrimSpace(sanitizeHeader(msg.Subject))
	msg.Message = strings.TrimSpace(msg.Message)

	if msg.Name == "" {
		return errors.New("name is required")
	}
	if len(msg.Name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}

	if len(msg.LastName) > 100 {
		return errors.New("lastname cannot exceed 100 characters")
	}

	if msg.Email == "" {
		return errors.New("email is required")
	}
	if len(msg.Email) > 150 {
		return errors.New("email cannot exceed 150 characters")
	}

	// Validate email format
	if _, err := mail.ParseAddress(msg.Email); err != nil || !emailRegex.MatchString(msg.Email) {
		return errors.New("invalid email format")
	}

	if len(msg.Subject) > 200 {
		return errors.New("subject cannot exceed 200 characters")
	}

	if msg.Message == "" {
		return errors.New("message is required")
	}
	if len(msg.Message) > 5000 {
		return errors.New("message content cannot exceed 5000 characters")
	}

	return nil
}

// sanitizeHeader strips CR and LF characters to prevent Email Header Injection
func sanitizeHeader(str string) string {
	str = strings.ReplaceAll(str, "\r", "")
	str = strings.ReplaceAll(str, "\n", "")
	return str
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
