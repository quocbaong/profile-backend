package repository

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"gorm.io/gorm"
)

type SkillRepository interface {
	FindAll(onlyActive bool) ([]models.SkillItem, error)
	FindByID(id uint) (*models.SkillItem, error)
	Create(skill *models.SkillItem) error
	Update(skill *models.SkillItem) error
	Delete(id uint) error
}

type skillRepository struct {
	db *gorm.DB
}

func NewSkillRepository(db *gorm.DB) SkillRepository {
	return &skillRepository{db: db}
}

func (r *skillRepository) FindAll(onlyActive bool) ([]models.SkillItem, error) {
	var skills []models.SkillItem
	query := r.db.Order("sort_order asc, id desc")
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&skills).Error; err != nil {
		return nil, err
	}
	return skills, nil
}

func (r *skillRepository) FindByID(id uint) (*models.SkillItem, error) {
	var skill models.SkillItem
	if err := r.db.First(&skill, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &skill, nil
}

func (r *skillRepository) Create(skill *models.SkillItem) error {
	return r.db.Create(skill).Error
}

func (r *skillRepository) Update(skill *models.SkillItem) error {
	return r.db.Save(skill).Error
}

func (r *skillRepository) Delete(id uint) error {
	return r.db.Delete(&models.SkillItem{}, id).Error
}
