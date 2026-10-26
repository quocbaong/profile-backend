package repository

import (
	"errors"

	"github.com/quocbaong/profile-backend/internal/models"
	"gorm.io/gorm"
)

type ProfileRepository interface {
	Get() (*models.ProfileSettings, error)
	Update(profile *models.ProfileSettings) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) Get() (*models.ProfileSettings, error) {
	var profile models.ProfileSettings
	if err := r.db.First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) Update(profile *models.ProfileSettings) error {
	var count int64
	r.db.Model(&models.ProfileSettings{}).Count(&count)
	if count == 0 {
		return r.db.Create(profile).Error
	}
	// Always update row 1
	var existing models.ProfileSettings
	if err := r.db.First(&existing).Error; err == nil {
		profile.ID = existing.ID
	}
	return r.db.Save(profile).Error
}
