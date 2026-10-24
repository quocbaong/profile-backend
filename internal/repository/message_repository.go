package repository

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"gorm.io/gorm"
)

type MessageRepository interface {
	FindAll() ([]models.ContactMessage, error)
	FindByID(id uint) (*models.ContactMessage, error)
	Create(msg *models.ContactMessage) error
	MarkAsRead(id uint) error
	Delete(id uint) error
	CountUnread() (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) FindAll() ([]models.ContactMessage, error) {
	var messages []models.ContactMessage
	if err := r.db.Order("created_at desc").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) FindByID(id uint) (*models.ContactMessage, error) {
	var msg models.ContactMessage
	if err := r.db.First(&msg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) Create(msg *models.ContactMessage) error {
	return r.db.Create(msg).Error
}

func (r *messageRepository) MarkAsRead(id uint) error {
	return r.db.Model(&models.ContactMessage{}).Where("id = ?", id).Update("is_read", true).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&models.ContactMessage{}, id).Error
}

func (r *messageRepository) CountUnread() (int64, error) {
	var count int64
	err := r.db.Model(&models.ContactMessage{}).Where("is_read = ?", false).Count(&count).Error
	return count, err
}
