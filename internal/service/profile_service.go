package service

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"github.com/quocbaong/profile-backend/internal/repository"
)

type ProfileService interface {
	Get() (*models.ProfileSettings, error)
	Update(profile *models.ProfileSettings) error
}

type profileService struct {
	repo repository.ProfileRepository
}

func NewProfileService(repo repository.ProfileRepository) ProfileService {
	return &profileService{repo: repo}
}

func (s *profileService) Get() (*models.ProfileSettings, error) {
	profile, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.New("profile settings not found")
	}
	return profile, nil
}

func (s *profileService) Update(profile *models.ProfileSettings) error {
	if profile.Name == "" {
		return errors.New("name is required")
	}
	return s.repo.Update(profile)
}
