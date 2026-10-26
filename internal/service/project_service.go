package service

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/repository"
)

type ProjectService interface {
	GetAll(onlyActive bool) ([]models.Project, error)
	GetByID(id uint) (*models.Project, error)
	Create(project *models.Project) error
	Update(id uint, project *models.Project) (*models.Project, error)
	Delete(id uint) error
}

type projectService struct {
	repo repository.ProjectRepository
}

func NewProjectService(repo repository.ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) GetAll(onlyActive bool) ([]models.Project, error) {
	return s.repo.FindAll(onlyActive)
}

func (s *projectService) GetByID(id uint) (*models.Project, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	return project, nil
}

func (s *projectService) Create(project *models.Project) error {
	if project.Title == "" {
		return errors.New("title is required")
	}
	return s.repo.Create(project)
}

func (s *projectService) Update(id uint, input *models.Project) (*models.Project, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("project not found")
	}

	existing.Title = input.Title
	existing.Category = input.Category
	existing.Description = input.Description
	existing.Tags = input.Tags
	existing.Image = input.Image
	existing.LiveURL = input.LiveURL
	existing.GithubURL = input.GithubURL
	existing.SortOrder = input.SortOrder
	existing.IsActive = input.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *projectService) Delete(id uint) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("project not found")
	}
	return s.repo.Delete(id)
}
