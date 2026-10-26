package service

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/repository"
)

type SkillService interface {
	GetAll(onlyActive bool) ([]models.SkillItem, error)
	GetByID(id uint) (*models.SkillItem, error)
	Create(skill *models.SkillItem) error
	Update(id uint, skill *models.SkillItem) (*models.SkillItem, error)
	Delete(id uint) error
}

type skillService struct {
	repo repository.SkillRepository
}

func NewSkillService(repo repository.SkillRepository) SkillService {
	return &skillService{repo: repo}
}

func (s *skillService) GetAll(onlyActive bool) ([]models.SkillItem, error) {
	return s.repo.FindAll(onlyActive)
}

func (s *skillService) GetByID(id uint) (*models.SkillItem, error) {
	skill, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if skill == nil {
		return nil, errors.New("skill not found")
	}
	return skill, nil
}

func (s *skillService) Create(skill *models.SkillItem) error {
	if skill.Name == "" {
		return errors.New("skill name is required")
	}
	return s.repo.Create(skill)
}

func (s *skillService) Update(id uint, input *models.SkillItem) (*models.SkillItem, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("skill not found")
	}

	existing.Name = input.Name
	existing.Category = input.Category
	existing.Icon = input.Icon
	existing.Color = input.Color
	existing.SortOrder = input.SortOrder
	existing.IsActive = input.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *skillService) Delete(id uint) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("skill not found")
	}
	return s.repo.Delete(id)
}
