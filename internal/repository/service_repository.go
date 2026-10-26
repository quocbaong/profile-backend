package repository

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"gorm.io/gorm"
)

type ServiceRepository interface {
	FindAll(onlyActive bool) ([]models.ServiceItem, error)
	FindByID(id uint) (*models.ServiceItem, error)
	Create(service *models.ServiceItem) error
	Update(service *models.ServiceItem) error
	Delete(id uint) error
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) FindAll(onlyActive bool) ([]models.ServiceItem, error) {
	var services []models.ServiceItem
	query := r.db.Order("sort_order asc, id desc")
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

func (r *serviceRepository) FindByID(id uint) (*models.ServiceItem, error) {
	var service models.ServiceItem
	if err := r.db.First(&service, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) Create(service *models.ServiceItem) error {
	return r.db.Create(service).Error
}

func (r *serviceRepository) Update(service *models.ServiceItem) error {
	return r.db.Save(service).Error
}

func (r *serviceRepository) Delete(id uint) error {
	return r.db.Delete(&models.ServiceItem{}, id).Error
}
