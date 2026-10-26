package service

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/repository"
)

type ServiceService interface {
	GetAll(onlyActive bool) ([]models.ServiceItem, error)
	GetByID(id uint) (*models.ServiceItem, error)
	Create(service *models.ServiceItem) error
	Update(id uint, service *models.ServiceItem) (*models.ServiceItem, error)
	Delete(id uint) error
}

type serviceService struct {
	repo repository.ServiceRepository
}

func NewServiceService(repo repository.ServiceRepository) ServiceService {
	return &serviceService{repo: repo}
}

func (s *serviceService) GetAll(onlyActive bool) ([]models.ServiceItem, error) {
	return s.repo.FindAll(onlyActive)
}

func (s *serviceService) GetByID(id uint) (*models.ServiceItem, error) {
	srv, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, errors.New("service not found")
	}
	return srv, nil
}

func (s *serviceService) Create(srv *models.ServiceItem) error {
	if srv.Title == "" {
		return errors.New("title is required")
	}
	return s.repo.Create(srv)
}

func (s *serviceService) Update(id uint, input *models.ServiceItem) (*models.ServiceItem, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("service not found")
	}

	existing.Title = input.Title
	existing.IconClass = input.IconClass
	existing.Description = input.Description
	existing.SortOrder = input.SortOrder
	existing.IsActive = input.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *serviceService) Delete(id uint) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("service not found")
	}
	return s.repo.Delete(id)
}
